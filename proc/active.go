package proc

import (
	"os"
	"path/filepath"

	"ctrz/cgroup"
)

func IsProcActive(pid int, starttime uint64) bool {
	procStats, err := ProcessStats(pid)
	if err != nil {
		return false
	}

	if procStats.Starttime != starttime {
		return false
	}

	path, err := cgroup.PathForPID(pid)
	if err != nil {
		return false
	}

	if _, err := os.Stat(filepath.Join(path, "cgroup.stat")); err != nil {
		return false
	}

	switch procStats.State {
	case 'T', 'Z', 'X', 'x':
		return false
	}

	return true
}
