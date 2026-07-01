//go:build linux
// +build linux

package main

import (
	"ctrz/cmd"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "__ctrz_init" {
		ctrzInit()
		return
	}

	cmd.Execute()
}