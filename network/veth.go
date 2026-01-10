package network

import (
	"log"
	"os/exec"
	"strconv"
)

func SetupVeth(pid int) error {
	hostIf := "veth-host-" + strconv.Itoa(pid)
	ctrzIf := "ctrz0"

	cmds := [][]string{
		{"ip", "link", "add", hostIf, "type", "veth", "peer", "name", ctrzIf},
		{"ip", "link", "set", ctrzIf, "netns", strconv.Itoa(pid)},
		{"ip", "addr", "add", "10.200.1.1/24", "dev", hostIf},
		{"ip", "link", "set", hostIf, "up"},
	}

	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		if err := cmd.Run(); err != nil {
			log.Fatal(cmd.String())
			return err
		}
	}
	return nil
}
