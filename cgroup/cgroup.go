//go:build linux

package cgroup

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/opencontainers/cgroups"
	cgroupmanager "github.com/opencontainers/cgroups/manager"
)

type Conf struct {
	Quota  int64
	Period uint64
	Memory int64
	CPU    string
}

type cgroupController interface {
	Apply(pid int) error
	Set(r *cgroups.Resources) error
	AddPid(subsystem string, pid int) error
	Destroy() error
}

var _ cgroupController = cgroups.Manager(nil)

func (m Manager) CreateAndAttach(conf *Conf, pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid %d", pid)
	}

	resources, err := resourcesFromConf(conf)
	if err != nil {
		return err
	}

	cm, err := cgroupmanager.NewWithPaths(
		&cgroups.Cgroup{Resources: resources},
		map[string]string{"": m.cgroupPath(pid)},
	)
	if err != nil {
		return fmt.Errorf("creating cgroup manager: %w", err)
	}

	return applyAndAttach(cm, resources, pid)
}

func (m Manager) cgroupPath(pid int) string {
	return filepath.Join(m.Root, fmt.Sprintf("ctrz-%d", pid))
}

func resourcesFromConf(conf *Conf) (*cgroups.Resources, error) {
	if conf == nil {
		return nil, errors.New("cgroup conf is nil")
	}

	quota, period, err := parseCPUMax(conf.CPU)
	if err != nil {
		return nil, err
	}

	return &cgroups.Resources{
		CpuQuota:  quota,
		CpuPeriod: period,
		Memory:    conf.Memory,
	}, nil
}

func applyAndAttach(cm cgroupController, res *cgroups.Resources, pid int) (err error) {
	defer func() {
		if err == nil {
			return
		}
		if derr := cm.Destroy(); derr != nil {
			err = errors.Join(err, fmt.Errorf("removing cgroup after failure: %w", derr))
		}
	}()

	if err = cm.Apply(-1); err != nil {
		return fmt.Errorf("creating cgroup: %w", err)
	}

	if err = cm.Set(res); err != nil {
		return fmt.Errorf("setting cgroup resources: %w", err)
	}

	if err = cm.AddPid("", pid); err != nil {
		return fmt.Errorf("attaching process %d: %w", pid, err)
	}

	return nil
}

// TODO: make this function obsolete. `ctrz run` builds a cpu string that is deconstructed here. Use a cpu struct instead
func parseCPUMax(value string) (int64, uint64, error) {
	fields := strings.Fields(value)
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("Invalid cpu.max value: %q\n", value)
	}

	var quota int64

	parsed, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || parsed <= 0 {
		return 0, 0, fmt.Errorf("Invalid CPU quota: %q\n", fields[0])
	}
	quota = parsed

	period, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil || period == 0 {
		return 0, 0, fmt.Errorf("Invalid CPU period: %q\n", fields[1])
	}

	return quota, period, nil
}
