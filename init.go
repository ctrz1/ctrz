//go:build linux

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"ctrz/fs"
	"ctrz/mount"
	"ctrz/network"
	"ctrz/utils"
)

func ctrzInit() {
	cont, err := utils.ListenToSocket()
	if err != nil {
		log.Fatalf("Error listening to socket: %v\n", err)
	}

	containerIP := cont.IP
	containerID := cont.ID
	cmd := os.Args[2]
	args := os.Args[2:]

	if err := syscall.Kill(1, syscall.SIGSTOP); err != nil {
		log.Fatal(err)
	}

	rootfs, err := fs.MountRootFs(containerID)
	if err != nil {
		log.Fatalf("Error mounting rootfs: %v\n", err)
	}

	cmdPath, err := fs.InjectBinary(cmd, fmt.Sprintf("%s/app", rootfs))
	if err != nil {
		log.Fatalf("Error injecting %s into namespace: %v\n", cmd, err)
	}

	if err := mount.PrepHost(rootfs); err != nil {
		log.Fatal(err)
	}

	if err := syscall.Sethostname([]byte(filepath.Base(cmd))); err != nil {
		log.Fatalf("Error setting new hostname: %v\n", err)
	}

	if err := os.MkdirAll("/root", 0o755); err != nil {
		log.Fatalf("Error creating user home /root: %v\n", err)
	}

	if err := mount.Proc(); err != nil {
		log.Fatal(err)
	}

	if err := mount.Dev(); err != nil {
		fmt.Print(err)
	}

	if err := mount.Etc(); err != nil {
		fmt.Print(err)
	}

	manager, err := network.New("10.200.1.0/24", "ctrz-br0", "ctrz0")
	if err != nil {
		log.Fatal(err)
	}

	if err := manager.SetupNetns(containerIP, cont.ExposePorts); err != nil {
		log.Fatal("run failed: ", err)
	}

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "no command specified")
		os.Exit(1)
	}

	env, err := mount.Environment(cont.Env)
	if err != nil {
		log.Fatalf("Error setting environment: %v\n", err)
	}

	err = syscall.Exec(cmdPath, args, env)
	if err != nil {
		log.Fatal("exec failed: ", err)
	}
}
