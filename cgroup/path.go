package cgroup

import "fmt"

func PathForPID(pid int) (string, error) {
	if pid < 0 {
		return "", fmt.Errorf("Invalid PID: %d\n", pid)
	}
	return fmt.Sprintf("/sys/fs/cgroup/ctrz-%d", pid), nil
}
