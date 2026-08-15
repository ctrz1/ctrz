package proc

func IsProcActive(pid int, starttime uint64) bool {
	procStats, err := ProcessStats(pid)
	if err != nil {
		return false
	}

	if procStats.Starttime != starttime {
		return false
	}

	switch procStats.State {
	case 'T', 'Z', 'X', 'x':
		return false
	}

	return true
}
