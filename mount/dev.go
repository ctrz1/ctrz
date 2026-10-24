package mount

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func Dev() error {
	if err := os.MkdirAll("/dev", 0o755); err != nil {
		log.Fatalf("Error creating /dev directory: %v\n", err)
	}

	if err := syscall.Mknod("/dev/null", syscall.S_IFCHR|0o666, int(unix.Mkdev(1, 3))); err != nil {
		return fmt.Errorf("Error creating /dev/null: %v\n", err)
	}
	if err := syscall.Mknod("/dev/zero", syscall.S_IFCHR|0o666, int(unix.Mkdev(1, 5))); err != nil {
		return fmt.Errorf("Error creating /dev/zero: %v\n", err)
	}
	if err := syscall.Mknod("/dev/random", syscall.S_IFCHR|0o666, int(unix.Mkdev(1, 8))); err != nil {
		return fmt.Errorf("Error creating /dev/random: %v\n", err)
	}
	if err := syscall.Mknod("/dev/urandom", syscall.S_IFCHR|0o666, int(unix.Mkdev(1, 9))); err != nil {
		return fmt.Errorf("Error creating /dev/urandom: %v\n", err)
	}

	// link stdin, stdout, stderr here
	return nil
}
