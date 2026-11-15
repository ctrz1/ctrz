//go:build linux

package cgroup

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/opencontainers/cgroups"
)

// fakeController records every call and can be told to fail at a given step.
type fakeController struct {
	calls []string

	applyErr   error
	setErr     error
	addPidErr  error
	destroyErr error

	gotResources *cgroups.Resources
}

func (f *fakeController) Apply(pid int) error {
	f.calls = append(f.calls, "Apply("+strconv.Itoa(pid)+")")
	return f.applyErr
}

func (f *fakeController) Set(r *cgroups.Resources) error {
	f.calls = append(f.calls, "Set")
	f.gotResources = r
	return f.setErr
}

func (f *fakeController) AddPid(subsystem string, pid int) error {
	f.calls = append(f.calls, "AddPid("+subsystem+","+strconv.Itoa(pid)+")")
	return f.addPidErr
}

func (f *fakeController) Destroy() error {
	f.calls = append(f.calls, "Destroy")
	return f.destroyErr
}

func TestApplyAndAttach_Success(t *testing.T) {
	f := &fakeController{}
	res := &cgroups.Resources{CpuQuota: 50000, CpuPeriod: 100000, Memory: 64 << 20}

	if err := applyAndAttach(f, res, 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Order matters: the cgroup must exist before limits are set, and the
	// limits must be in place before the process is moved in.
	want := []string{"Apply(-1)", "Set", `AddPid(,42)`}
	if !reflect.DeepEqual(f.calls, want) {
		t.Errorf("calls = %v, want %v", f.calls, want)
	}
	if f.gotResources != res {
		t.Errorf("Set got %+v, want the resources that were passed in", f.gotResources)
	}
}

func TestApplyAndAttach_FailureRemovesCgroup(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name      string
		fake      *fakeController
		wantCalls []string
		wantInMsg string
	}{
		{
			name:      "apply fails",
			fake:      &fakeController{applyErr: boom},
			wantCalls: []string{"Apply(-1)", "Destroy"},
			wantInMsg: "creating cgroup",
		},
		{
			name:      "set fails",
			fake:      &fakeController{setErr: boom},
			wantCalls: []string{"Apply(-1)", "Set", "Destroy"},
			wantInMsg: "setting cgroup resources",
		},
		{
			name:      "add pid fails",
			fake:      &fakeController{addPidErr: boom},
			wantCalls: []string{"Apply(-1)", "Set", "AddPid(,7)", "Destroy"},
			wantInMsg: "attaching process 7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyAndAttach(tt.fake, &cgroups.Resources{}, 7)

			if !errors.Is(err, boom) {
				t.Fatalf("error %v does not wrap the underlying error", err)
			}
			if !strings.Contains(err.Error(), tt.wantInMsg) {
				t.Errorf("error %q does not mention %q", err, tt.wantInMsg)
			}
			if !reflect.DeepEqual(tt.fake.calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", tt.fake.calls, tt.wantCalls)
			}
		})
	}
}

func TestApplyAndAttach_FailedCleanupIsReported(t *testing.T) {
	setErr := errors.New("set failed")
	destroyErr := errors.New("destroy failed")
	f := &fakeController{setErr: setErr, destroyErr: destroyErr}

	err := applyAndAttach(f, &cgroups.Resources{}, 1)

	if !errors.Is(err, setErr) {
		t.Errorf("original error lost: %v", err)
	}
	if !errors.Is(err, destroyErr) {
		t.Errorf("cleanup error lost: %v", err)
	}
}

func TestResourcesFromConf(t *testing.T) {
	got, err := resourcesFromConf(&Conf{CPU: "50000 100000", Memory: 64 << 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := &cgroups.Resources{CpuQuota: 50000, CpuPeriod: 100000, Memory: 64 << 20}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestResourcesFromConf_Invalid(t *testing.T) {
	if _, err := resourcesFromConf(nil); err == nil {
		t.Error("nil conf: expected error")
	}
	if res, err := resourcesFromConf(&Conf{CPU: "garbage"}); err == nil || res != nil {
		t.Errorf("invalid cpu: got (%v, %v), want (nil, error)", res, err)
	}
}

func TestManagerCgroupPath(t *testing.T) {
	tests := []struct {
		root string
		pid  int
		want string
	}{
		{"/sys/fs/cgroup", 42, "/sys/fs/cgroup/ctrz-42"},
		{"/sys/fs/cgroup/", 42, "/sys/fs/cgroup/ctrz-42"},
		{"/sys/fs/cgroup/ctrz", 1234567, "/sys/fs/cgroup/ctrz/ctrz-1234567"},
	}

	for _, tt := range tests {
		if got := (Manager{Root: tt.root}).cgroupPath(tt.pid); got != tt.want {
			t.Errorf("Root=%q pid=%d: got %q, want %q", tt.root, tt.pid, got, tt.want)
		}
	}
}

// Invalid input must be rejected before anything is created in the cgroup
// filesystem. Root points at an empty temp dir, so any side effect shows up.
func TestCreateAndAttach_RejectsInvalidInput(t *testing.T) {
	valid := &Conf{CPU: "50000 100000", Memory: 1 << 20}

	tests := []struct {
		name string
		conf *Conf
		pid  int
	}{
		{"nil conf", nil, 42},
		{"pid zero would attach the caller", valid, 0},
		{"negative pid", valid, -1},
		{"invalid cpu string", &Conf{CPU: "max 100000"}, 42},
		{"empty cpu string", &Conf{}, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()

			if err := (Manager{Root: root}).CreateAndAttach(tt.conf, tt.pid); err == nil {
				t.Fatal("expected error, got nil")
			}

			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Errorf("rejected call still created %d entries under Root", len(entries))
			}
		})
	}
}

func TestParseCPUMax(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantQuota  int64
		wantPeriod uint64
		wantErr    bool
	}{
		{
			name:       "valid CPU limit",
			input:      "50000 100000",
			wantQuota:  50000,
			wantPeriod: 100000,
		},
		{
			name:    "missing quota",
			input:   "100000",
			wantErr: true,
		},
		{
			name:    "too many fields",
			input:   "50000 100000 200000",
			wantErr: true,
		},
		{
			name:    "invalid quota",
			input:   "abc 100000",
			wantErr: true,
		},
		{
			name:    "negative quota",
			input:   "-50000 100000",
			wantErr: true,
		},
		{
			name:    "zero quota",
			input:   "0 100000",
			wantErr: true,
		},
		{
			name:    "invalid period",
			input:   "50000 abc",
			wantErr: true,
		},
		{
			name:    "zero period",
			input:   "50000 0",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace",
			input:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quota, period, err := parseCPUMax(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if quota != tt.wantQuota {
				t.Errorf("Got quota = %d, want %d",
					quota, tt.wantQuota)
			}

			if period != tt.wantPeriod {
				t.Errorf("Got period = %d, want %d",
					period, tt.wantPeriod)
			}
		})
	}
}
