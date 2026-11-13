//go:build linux

package mount

import (
	"fmt"
	"os"
	"syscall"
)

func Proc() error {
	if err := os.MkdirAll("/proc", 0o555); err != nil {
		return fmt.Errorf("Error creating '/proc: %v\n", err)
	}

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("Error mounting pseudo filesystem: %v\n", err)
	}
	return nil
}
