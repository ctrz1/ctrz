//go:build linux

package mount

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func PrepHost(rootfs string) error {
	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("Error preventing mount propagation: %v\n", err)
	}

	if err := syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Error ensuring rootfs is a mount point: %v\n", err)
	}

	old := filepath.Join(rootfs, ".oldroot")
	if err := os.MkdirAll(old, 0o700); err != nil {
		return fmt.Errorf("Error creating '.oldroot' dir: %v\n", err)
	}

	if err := syscall.PivotRoot(rootfs, old); err != nil {
		return fmt.Errorf("Error pivoting into container rootfs: %v\n", err)
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("Error moving to new root: %v\n", err)
	}

	if err := syscall.Unmount("/.oldroot", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("Error unmounting old root: %v\n", err)
	}
	if err := os.RemoveAll("/.oldroot"); err != nil {
		return fmt.Errorf("Error deleting old root: %v\n", err)
	}
	return nil
}
