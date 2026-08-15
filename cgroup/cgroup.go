package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func (m Manager) CreateAndAttach(pid int, cpuMax string) error {
	path := filepath.Join(m.Root, fmt.Sprintf("ctrz-%d", pid))
	if err := os.Mkdir(path, 0o755); err != nil && !os.IsExist(err) {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, "cpu.max"), []byte(cpuMax), 0o644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, "cgroup.procs"),
		[]byte(strconv.Itoa(pid)), 0o644); err != nil {
		return err
	}

	return nil
}
