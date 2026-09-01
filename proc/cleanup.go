package proc

import (
	"ctrz/spec"
	"fmt"
	"syscall"
)

func Kill(rm *spec.Removal, pid int, starttime uint64) error {
	if IsProcActive(pid, starttime) {
		fmt.Printf("Killing process %d\n", pid)
		if rm.Force {
			if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
				return fmt.Errorf("Error killing container: %v\n", err)
			}
		} else {
			if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
				return fmt.Errorf("Error killing container: %v\n", err)
			}
		}
	}
	return nil
}
