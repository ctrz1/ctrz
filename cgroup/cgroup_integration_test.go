//go:build linux && integration

package cgroup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const cgroupV2Mount = "/sys/fs/cgroup"

func requireCgroupV2(t *testing.T) {
	t.Helper()
	if os.Geteuid() != 0 {
		t.Skip("needs root")
	}
	if _, err := os.Stat(filepath.Join(cgroupV2Mount, "cgroup.controllers")); err != nil {
		t.Skipf("no cgroup v2 mount at %s", cgroupV2Mount)
	}
}

func testRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(cgroupV2Mount, fmt.Sprintf("ctrz-test-%d", os.Getpid()))
	t.Cleanup(func() { _ = os.Remove(root) })
	return root
}

func readTrimmed(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

func TestCreateAndAttach_Integration(t *testing.T) {
	requireCgroupV2(t)
	root := testRoot(t)

	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pid := cmd.Process.Pid
	cgroupDir := filepath.Join(root, fmt.Sprintf("ctrz-%d", pid))
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		_ = os.Remove(cgroupDir)
	})

	conf := &Conf{CPU: "50000 100000", Memory: 64 << 20}
	if err := (Manager{Root: root}).CreateAndAttach(conf, pid); err != nil {
		t.Fatalf("CreateAndAttach: %v", err)
	}

	if got := readTrimmed(t, filepath.Join(cgroupDir, "cpu.max")); got != "50000 100000" {
		t.Errorf("cpu.max = %q, want %q", got, "50000 100000")
	}
	if got := readTrimmed(t, filepath.Join(cgroupDir, "memory.max")); got != strconv.Itoa(64<<20) {
		t.Errorf("memory.max = %q, want %d", got, 64<<20)
	}

	procs := strings.Fields(readTrimmed(t, filepath.Join(cgroupDir, "cgroup.procs")))
	if len(procs) != 1 || procs[0] != strconv.Itoa(pid) {
		t.Errorf("cgroup.procs = %v, want [%d]", procs, pid)
	}
}

// If the process cannot be attached, the freshly created cgroup must not be
// left behind.
func TestCreateAndAttach_Integration_CleansUpOnFailure(t *testing.T) {
	requireCgroupV2(t)
	root := testRoot(t)

	// A pid that is guaranteed to be gone: start a process and reap it.
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	deadPid := cmd.Process.Pid

	conf := &Conf{CPU: "50000 100000", Memory: 64 << 20}
	if err := (Manager{Root: root}).CreateAndAttach(conf, deadPid); err == nil {
		t.Fatal("expected error when attaching a non-existent process")
	}

	cgroupDir := filepath.Join(root, fmt.Sprintf("ctrz-%d", deadPid))
	if _, err := os.Stat(cgroupDir); !os.IsNotExist(err) {
		t.Errorf("cgroup %s was left behind (stat err: %v)", cgroupDir, err)
	}
}
