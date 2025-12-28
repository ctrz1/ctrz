package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateAndAttach(pid int, cpuMax string) error {
	group := PathForPID(pid)

	if err := os.Mkdir(group, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	if err := os.WriteFile(filepath.Join(group, "cpu.max"), []byte(cpuMax), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(group, "cgroup.procs"),
		[]byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return err
	}

	return nil
}
