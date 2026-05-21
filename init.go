package main

import (
	"ctrz/network"
	"fmt"
	"log"
	"os"
	"syscall"
)

func ctrzInit() {
	if err := syscall.Kill(os.Getpid(), syscall.SIGSTOP); err != nil {
		log.Fatal(err)
	}
	err := network.SetupNetns("10.200.1.2")
	if err != nil {
		log.Fatal("run failed: ", err)
	}

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
