//go:build linux

package main

import (
	"log"
	"os"

	"ctrz/cmd"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "__ctrz_init" {
		ctrzInit()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "__deamon" {
		ctrzDeamon()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "__dashboard" {
		if err := StartDashboardServer(); err != nil {
			log.Fatal(err)
		}
		return
	}

	cmd.Execute()
}
