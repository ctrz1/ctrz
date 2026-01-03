package main

import (
	"ctrz/cmd"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "__ctrz_init" {
		ctrzInit()
		return
	}

	cmd.Execute();
}

func ctrzInit() {
	err := exec.Command("ip", "link", "set", "lo", "up").Run()
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
		os.Exit(1)
	}
}