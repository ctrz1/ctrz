package cgroup

import (
	"fmt"
	"path/filepath"
)

func (m Manager) Path(pid int) (string, error) {
	if pid < 0 {
		return "", fmt.Errorf("Invalid PID: %d\n", pid)
	}
	return filepath.Join(m.Root, fmt.Sprintf("ctrz-%d", pid)), nil
}
