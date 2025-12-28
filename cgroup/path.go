package cgroup

import "fmt"

func PathForPID(pid int) string {
	return fmt.Sprintf("/sys/fs/cgroup/ctrz-%d", pid)
}
