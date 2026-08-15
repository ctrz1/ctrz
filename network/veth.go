//go:build linux

package network

import (
	"fmt"
	"os/exec"
	"strconv"
)

func (m Manager) SetupVeth(pid int) error {
	hostIf := "veth-host-" + strconv.Itoa(pid)
	ctrzIf := "ctrz0"

	cmds := [][]string{
		{"ip", "link", "add", hostIf, "type", "veth", "peer", "name", ctrzIf},
		{"ip", "link", "set", ctrzIf, "netns", strconv.Itoa(pid)},
		{"ip", "link", "set", hostIf, "master", "ctrz-br0"},
		{"ip", "link", "set", hostIf, "up"},
	}

	for _, c := range cmds {
		out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%v: %s", err, out)
		}
	}
	return nil
}
