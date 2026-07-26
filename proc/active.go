package proc

import (
	"os"
	"path/filepath"

	"ctrz/cgroup"
)

func IsProcActive(pid int) bool {
	path, err := cgroup.PathForPID(pid)
	if err != nil {
		return false
	}
	if _, err = os.Stat(filepath.Join(path, "cgroup.stat")); err != nil {
		return false
	}
	return true
}
