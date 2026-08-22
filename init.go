//go:build linux

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"ctrz/fs"
	"ctrz/network"

	"golang.org/x/sys/unix"
)

func ctrzInit() {
	containerIP := os.Args[2]
	containerID := os.Args[3]
	cmd := os.Args[4]
	args := os.Args[4:]

	if err := syscall.Kill(1, syscall.SIGSTOP); err != nil {
		log.Fatal(err)
	}

	rootfs, err := fs.MountRootFs(containerID)
	if err != nil {
		log.Fatalf("Error mounting rootfs: %v\n", err)
	}

	if err := fs.InjectBinary(cmd, fmt.Sprintf("%s/app", rootfs)); err != nil {
		log.Fatalf("Error injecting %s into namespace: %v\n", cmd, err)
	}

	// 1. prevent mount propagation to host
	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		log.Fatalf("Error preventing mount propagation: %v\n", err)
	}

	// 2. ensure rootfs is a mount point
	if err := syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		log.Fatalf("Error ensuring rootfs is a mount point: %v\n", err)
	}

	// 3. prepare pivot_root
	old := filepath.Join(rootfs, ".oldroot")
	if err := os.MkdirAll(old, 0o700); err != nil {
		log.Fatalf("Error creating '.oldroot' dir: %v\n", err)
	}

	// 4. pivot into container rootfs
	if err := syscall.PivotRoot(rootfs, old); err != nil {
		log.Fatalf("Error pivoting into container rootfs: %v\n", err)
	}

	// 5. move to new root
	if err := os.Chdir("/"); err != nil {
		log.Fatalf("Error moving to new root: %v\n", err)
	}

	// 6. unmount old host root
	if err := syscall.Unmount("/.oldroot", syscall.MNT_DETACH); err != nil {
		log.Fatalf("Error unmounting old root: %v\n", err)
	}
	if err := os.RemoveAll("/.oldroot"); err != nil {
		log.Fatalf("Error deleting old root: %v\n", err)
	}

	// 7. hostname (UTS namespace)
	if err := syscall.Sethostname([]byte("ctrz")); err != nil {
		log.Fatalf("Error setting new hostname: %v\n", err)
	}

	if err := os.MkdirAll("/dev", 0o755); err != nil {
		log.Fatalf("Error creating /dev directory: %v\n", err)
	}

	if err := syscall.Mknod("/dev/null", syscall.S_IFCHR|0o666, int(unix.Mkdev(1, 3))); err != nil {
		fmt.Printf("Error creating /dev/null: %v\n", err)
	}
	if err := syscall.Mknod("/dev/zero", syscall.S_IFCHR|0o666, int(unix.Mkdev(1, 5))); err != nil {
		fmt.Printf("Error creating /dev/zero: %v\n", err)
	}
	if err := syscall.Mknod("/dev/random", syscall.S_IFCHR|0o666, int(unix.Mkdev(1, 8))); err != nil {
		fmt.Printf("Error creating /dev/random: %v\n", err)
	}
	if err := syscall.Mknod("/dev/urandom", syscall.S_IFCHR|0o666, int(unix.Mkdev(1, 9))); err != nil {
		fmt.Printf("Error creating /dev/urandom: %v\n", err)
	}

	if err := os.MkdirAll("/proc", 0o555); err != nil {
		log.Fatalf("Error creating '/proc: %v\n", err)
	}

	// 8. mount pseudo filesystems
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		log.Fatalf("Error mounting pseudo filesystem: %v\n", err)
	}

	manager := network.New("10.200.1.0/24", "ctrz-br0", "ctrz0")

	if err := manager.SetupNetns(containerIP); err != nil {
		log.Fatal("run failed: ", err)
	}

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "no command specified")
		os.Exit(1)
	}

	err = syscall.Exec("/app/bin", args, os.Environ())
	if err != nil {
		log.Fatal("exec failed: ", err)
	}
}
