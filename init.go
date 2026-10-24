//go:build linux

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ctrz/fs"
	"ctrz/mount"
	"ctrz/network"
	"ctrz/spec"
	"ctrz/utils"
)

func ctrzInit() {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		log.Fatal(err)
	}
	sockPath := filepath.Join(dir, "ctrz.sock")
	cont, err := readFromSock(sockPath)
	if err != nil {
		log.Fatal(err)
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

func readFromSock(sockPath string) (cont spec.ContainerSocket, err error) {
	l, err := utils.NewSocketListener(sockPath)
	if err != nil {
		return spec.ContainerSocket{}, err
	}
	defer func() {
		err = errors.Join(err, l.Close())
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cont, err = l.ListenToSocket(rctx)
	if err != nil {
		return spec.ContainerSocket{}, fmt.Errorf("Error listening to socket: %v\n", err)
	}
	return cont, nil
}
