//go:build linux
// +build linux

package main

import (
	"ctrz/fs"
	"ctrz/network"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

func ctrzInit() {
	if err := syscall.Kill(1, syscall.SIGSTOP); err != nil {
		log.Fatal(err)
	}

	o, err := exec.Command("id").CombinedOutput()
	fmt.Printf("Current UID: %s, error: %v\n", string(o), err)

	rootfs, err := fs.MountRootFs("containerID")
	if err != nil {
		log.Fatalf("Error mounting rootfs: %v\n", err)
	}

	fmt.Printf("rootfs: %s\n", rootfs)

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
	if err := os.MkdirAll(old, 0700); err != nil {
		log.Fatalf("Error creating '.oldroot' dir: %v\n", err)
	}

	data, err := os.ReadFile("/proc/self/uid_map")
	fmt.Println(string(data), err)

	data, err = os.ReadFile("/proc/self/gid_map")
	fmt.Println(string(data), err)

	data, err = os.ReadFile("/proc/self/status")
	fmt.Println(string(data), err)

	// 4. pivot into container rootfs
	if err := syscall.PivotRoot(rootfs, old); err != nil {
		log.Fatalf("Error pivoting into container rootfs: %v\n", err)
	}

	fmt.Println("cwd:")
	cwd, err := os.Getwd()
	fmt.Println("Error:", err)
	fmt.Println(cwd)

	fmt.Println("root listing")
	entries, err := os.ReadDir("/")
	fmt.Println("Error:", err)
	for _, e := range entries {
	    fmt.Println(e.Name())
	}

	fmt.Println("bin listing")
	entries, err = os.ReadDir("/bin")
	fmt.Println("Error:", err)
	for _, e := range entries {
	    fmt.Println(e.Name())
	}

	st := syscall.Stat_t{}
	err = syscall.Stat("/", &st)
	fmt.Printf("root dev=%d inode=%d err=%v\n", st.Dev, st.Ino, err)

	err = os.WriteFile("/testfile", []byte("x"), 0644)
	fmt.Printf("write=%v\n", err)

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

	if err := os.MkdirAll("/dev", 0755); err != nil {
		log.Fatalf("Error creating /dev directory: %v\n", err)
	}

	if err != nil {
		log.Fatalf("Error mounting /dev: %v", err)
	}

	if err := syscall.Mknod("/dev/null", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 3))); err != nil {
		fmt.Printf("Error creating /dev/null: %v\n", err)
	}
	syscall.Mknod("/dev/zero", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 5)))
	syscall.Mknod("/dev/random", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 8)))
	syscall.Mknod("/dev/urandom", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 9)))

	if err := os.MkdirAll("/proc", 0555); err != nil {
		log.Fatalf("Error creating '/proc: %v\n", err)
	}

	data, err = os.ReadFile("/proc/self/status")
	if err != nil {
		fmt.Printf("Error redaing /proc/self/status: %v\n", err)
	}
	fmt.Println(string(data))

	fmt.Printf("uid=%d euid=%d gid=%d egid=%d\n",
		os.Getuid(),
		os.Geteuid(),
		os.Getgid(),
		os.Getegid(),
	)

	bin_path := os.Getenv("PATH")
	fmt.Printf("Current $PATH: %s\n", bin_path)


	fmt.Println(os.Stat("/bin"))
	fmt.Println(os.Stat("/bin/ls"))

	o, err = exec.Command("ls", "-lah").CombinedOutput()
	fmt.Printf("Current ls: %s, error: %v\n", string(o), err)

	// 8. mount pseudo filesystems
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		log.Fatalf("Error mounting pseudo filesystem: %v\n", err)
	}

	fmt.Println("pseudo filesystem mounted...")

	err = network.SetupNetns("10.200.1.2")
	if err != nil {
		log.Fatal("run failed: ", err)
	}

	fmt.Println("network st up...")

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "no command specified")
		os.Exit(1)
	}

	cmd := os.Args[2]
	args := os.Args[2:]

	err = syscall.Exec(cmd, args, os.Environ())
	if err != nil {
		log.Fatal("exec failed: ", err)
	}
}
