//go:build linux

package cgroup

import (
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

func (m Manager) CreateAndAttach(conf *Conf, pid int) error {
	quota, period, err := parseCPUMax(conf.CPU)
	if err != nil {
		return err
	}

	path := filepath.Join(
		m.Root,
		fmt.Sprintf("ctrz-%d", pid),
	)

	config := &cgroups.Cgroup{
		Resources: &cgroups.Resources{
			CpuQuota:  quota,
			CpuPeriod: period,
			Memory:    conf.Memory,
		},
	}

	cm, err := cgroupmanager.NewWithPaths(
		config,
		map[string]string{"": path},
	)
	if err != nil {
		return fmt.Errorf("Error creating cgroup manager: %w\n", err)
	}

	if err := cm.Apply(-1); err != nil {
		return fmt.Errorf("Error creating cgroup: %w\n", err)
	}

	if err := cm.Set(config.Resources); err != nil {
		return fmt.Errorf("Error setting cgroup resources: %w\n", err)
	}

	if err := cm.AddPid("", pid); err != nil {
		return fmt.Errorf("Error attaching process %d: %w\n", pid, err)
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
