//go:build linux

package network

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func (m Manager) SetupVeth(pid int) error {
	hostIf := "veth-host-" + strconv.Itoa(pid)
	ctrzIf := m.ContainerInterface

	host, err := netlink.LinkByName(m.Bridge)
	if err != nil {
		return fmt.Errorf("Error finding bridge %s: %v\n", m.Bridge, err)
	}

	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: hostIf,
		},
		PeerName: ctrzIf,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return fmt.Errorf("Error creating veth pair: %v\n", err)
	}
	hostLink, err := netlink.LinkByName(hostIf)
	if err != nil {
		return fmt.Errorf("Error finding host veth %s: %v\n", hostIf, err)
	}

	ctrLink, err := netlink.LinkByName(ctrzIf)
	if err != nil {
		return fmt.Errorf("Error finding container veth %s: %v\n", ctrzIf, err)
	}

	p := filepath.Join("/", "proc", strconv.Itoa(pid), "ns", "net")
	b, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("Namespace at %s does not exist: %v\n", p, err)
	}

	fd := netns.NsHandle(b.Fd())

	ns, err := netlink.NewHandleAt(fd)
	if err != nil {
		return fmt.Errorf("Error opening target netns: %v\n", err)
	}
	defer ns.Close()

	if err := netlink.LinkSetNsFd(ctrLink, int(fd)); err != nil {
		return fmt.Errorf("Error moving %s to netns: %v\n", ctrzIf, err)
	}

	if err := netlink.LinkSetMaster(hostLink, host); err != nil {
		return fmt.Errorf("Error attaching %s to bridge: %v\n", hostIf, err)
	}

	if err := netlink.LinkSetUp(hostLink); err != nil {
		return fmt.Errorf("Error bringing host veth up: %v\n", err)
	}
	return nil
}
