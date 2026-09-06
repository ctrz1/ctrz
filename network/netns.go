//go:build linux
// +build linux

package network

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"ctrz/spec"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"github.com/vishvananda/netlink"
)

/**
*	Here bridges, subnets, and container veths should not be hardcoded, but taken from the manager
*	It is acceptable for now though and should be addressed once these exec.Command(...) calls are
* 	replaced with netlink
**/

func (m Manager) SetupHostNetworking() error {
	if out, err := exec.Command(
		"sysctl",
		"-w",
		"net.ipv6.conf.all.forwarding=1",
	).CombinedOutput(); err != nil {
		return fmt.Errorf("enable ipv6_forwarding failed: %v: %s", err, out)
	}
	if out, err := exec.Command(
		"sysctl",
		"-w",
		"net.ipv4.ip_forward=1",
	).CombinedOutput(); err != nil {
		return fmt.Errorf("enable ip_forward failed: %v: %s", err, out)
	}
	//TODO: review this one (CVE-2020-8558)
	if out, err := exec.Command(
		"sysctl",
		"-w",
		"net.ipv4.conf.all.route_localnet=1",
	).CombinedOutput(); err != nil {
		return fmt.Errorf("enable ip_forward failed: %v: %s", err, out)
	}

	_, err := ensureBridge(m.Bridge, m.Gateway)
	if err != nil {
		return fmt.Errorf("Error ensuring bridge: %v\n", err)
	}

	if err := m.nat(); err != nil {
		return fmt.Errorf("Error establishing NAT rules: %v\n", err)
	}

	if err := m.outboundForward(); err != nil {
		return fmt.Errorf("Error establishing oubound rule: %v\n", err)
	}

	if err := m.allowReturnTraffic(); err != nil {
		return fmt.Errorf("Error allowing return traffic: %v\n", err)
	}

	return nil
}

func (m Manager) SetupNetns(containerIP string) error {
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("Error finding loopback: %v\n", err)
	}

	if err := netlink.LinkSetUp(lo); err != nil {
		return fmt.Errorf("Error bringing loopback up: %v\n", err)
	}

	link, err := netlink.LinkByName(m.ContainerInterface)
	if err != nil {
		return fmt.Errorf("Error finding %s: %v\n", m.ContainerInterface, err)
	}

	ip := net.ParseIP(containerIP)
	if ip == nil {
		return fmt.Errorf("Invalid container IP %s\n", containerIP)
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(24, 32),
		},
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("Error assigning container IP: %v\n", err)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("Error bringing container interface up: %v\n", err)
	}

	gateway := net.ParseIP(strings.Split(m.Gateway, "/")[0])
	if gateway == nil {
		return fmt.Errorf("Invalid gateway %s\n", m.Gateway)
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Gw:        gateway,
	}

	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("Error adding default route: %v\n", err)
	}

	return nil
}

func parsePorts(ports string) (spec.PortMapping, error) {
	parts := strings.Split(ports, ":")
	if len(parts) != 2 {
		return spec.PortMapping{}, fmt.Errorf("invalid port mapping: %s", ports)
	}
	hp, err := strconv.Atoi(parts[0])
	if err != nil {
		return spec.PortMapping{}, fmt.Errorf("Invalid host port: %s: %v", parts[0], err)
	}
	cp, err := strconv.Atoi(parts[1])
	if err != nil {
		return spec.PortMapping{}, fmt.Errorf("Invalid container port: %s: %v", parts[1], err)
	}
	if cp < 1 || hp < 1 || cp > 65535 || hp > 65535 {
		return spec.PortMapping{}, fmt.Errorf("invalid port mapping: %s", ports)
	}
	return spec.PortMapping{HostPort: hp, ContainerPort: cp}, nil
}

func DenyAllElse(containerIP string) error {
	cmds := [][]string{
		{
			"iptables", "-A", "FORWARD",
			"-d", containerIP,
			"-j", "DROP",
		},
	}

	for _, c := range cmds {
		out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%v: %s", err, out)
		}
	}

	return nil
}

func ensureBridge(name, gateway string) (netlink.Link, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		attrs := netlink.NewLinkAttrs()
		attrs.Name = name

		link = &netlink.Bridge{
			LinkAttrs: attrs,
		}

		if err := netlink.LinkAdd(link); err != nil {
			return nil, fmt.Errorf("create bridge %s: %v", name, err)
		}
	}

	if link.Type() != "bridge" {
		return nil, fmt.Errorf("%s exists but is not a bridge", name)
	}
	addr, err := netlink.ParseAddr(gateway)
	if err != nil {
		return nil, fmt.Errorf("parse bridge subnet %s: %v", gateway, err)
	}
	if err := netlink.AddrReplace(link, addr); err != nil {
		return nil, fmt.Errorf("configure bridge %s address %s: %v", name, gateway, err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return nil, fmt.Errorf("bring bridge %s up: %v", name, err)
	}

	return link, nil
}

func (m Manager) nat() error {
	c := m.Nftables.Conn
	table := m.Nftables.Table

	outboundRuleId := fmt.Sprintf("ctrz:nat:%s", m.Subnet)
	createOutbound := true
	inboundRuleId := "ctrz:nat:127.0.0.0/8"
	createInbound := true

	rules, err := c.GetRules(table, m.Nftables.Postruting)
	if err != nil {
		// TODO: think if this should really return an error. Worst case the same rule is created multiple times?
		return fmt.Errorf("Error determining if rules exist: %v\n", err)
	}
	for _, rule := range rules {
		if string(rule.UserData) == outboundRuleId {
			createOutbound = false
		}
		if string(rule.UserData) == inboundRuleId {
			createInbound = false
		}
	}

	if !(createOutbound || createInbound) {
		return nil
	}

	_, subnet, err := net.ParseCIDR(m.Subnet)
	if err != nil {
		return fmt.Errorf("Error parsing subnet %s: %v\n", m.Subnet, err)
	}

	subnetIP := subnet.IP.To4()
	if subnetIP == nil {
		return fmt.Errorf("Subnet %s is not IPv4\n", m.Subnet)
	}

	ones, bits := subnet.Mask.Size()
	if bits != 32 || ones != 24 {
		return fmt.Errorf("Subnet %s must be an IPv4 /24", m.Subnet)
	}

	// iptables -t nat -A POSTROUTING -s 10.200.1.0/24 -j MASQUERADE
	outbound := &nftables.Rule{
		Table:    table,
		Chain:    m.Nftables.Postruting,
		UserData: []byte(outboundRuleId),
		Exprs: []expr.Any{
			&expr.Payload{
				DestRegister: reg1,
				Base:         expr.PayloadBaseNetworkHeader,
				Len:          3,
				Offset:       12,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     subnetIP[:3],
			},
			&expr.Masq{},
		},
	}

	// iptables -t nat -A POSTROUTING -s 127.0.0.0/8 -d 10.200.1.0/24 -j MASQUERADE
	inbound := &nftables.Rule{
		Table:    table,
		Chain:    m.Nftables.Postruting,
		UserData: []byte(inboundRuleId),
		Exprs: []expr.Any{
			&expr.Payload{
				DestRegister: reg1,
				Base:         expr.PayloadBaseNetworkHeader,
				Len:          1,
				Offset:       12,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     []byte{127},
			},
			&expr.Payload{
				DestRegister: reg1,
				Base:         expr.PayloadBaseNetworkHeader,
				Len:          3,
				Offset:       16,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     subnetIP[:3],
			},
			&expr.Masq{},
		},
	}

	if createOutbound {
		c.AddRule(outbound)
	}
	if createInbound {
		c.AddRule(inbound)
	}

	return c.Flush()
}

func (m Manager) outboundForward() error {
	c := m.Nftables.Conn
	table := m.Nftables.Table
	ruleId := fmt.Sprintf("ctrz:forward:%s", m.Bridge)

	rules, err := c.GetRules(table, m.Nftables.Forward)
	if err != nil {
		return fmt.Errorf("Error determining if rule exists: %v\n", err)
	}

	for _, rule := range rules {
		if string(rule.UserData) == ruleId {
			return nil
		}
	}

	// iptables -A FORWARD -i ctrz-br0 -j ACCEPT
	// add rule ip ctrz forward iifname "ctrz-br0" accept
	c.AddRule(&nftables.Rule{
		Table:    table,
		Chain:    m.Nftables.Forward,
		UserData: []byte(ruleId),
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyIIFNAME,
				Register: reg1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     []byte(m.Bridge),
			},
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})

	return c.Flush()
}

func (m Manager) allowReturnTraffic() error {
	/*
		iptables -A FORWARD -o ctrz-br0 -m conntrack --cstate ESTABLISHED,RELATED -j ACCEPT
		sudo nft --debug=netlink add rule ip ctrz forward oifname "ctrz-br0" ct state established,related accept

		  [ meta load oifname => reg 1 ]
		  [ cmp eq reg 1 0x7a727463 0x3072622d 0x00000000 0x00000000 ]
		  [ ct load state => reg 1 ]
		  [ bitwise reg 1 = ( reg 1 & 0x00000006 ) ^ 0x00000000 ]
		  [ cmp neq reg 1 0x00000000 ]
		  [ immediate reg 0 accept ]

	*/
	c := m.Nftables.Conn
	table := m.Nftables.Table
	ruleId := fmt.Sprintf("ctrz:return:%s", m.Bridge)

	rules, err := c.GetRules(table, m.Nftables.Forward)
	if err != nil {
		return fmt.Errorf("Error determining if rule exists: %v\n", err)
	}

	for _, rule := range rules {
		if string(rule.UserData) == ruleId {
			return nil
		}
	}

	c.AddRule(&nftables.Rule{
		Table:    table,
		Chain:    m.Nftables.Forward,
		UserData: []byte(ruleId),
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyOIFNAME,
				Register: reg1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     []byte(m.Bridge),
			},
			&expr.Ct{
				Key:      expr.CtKeySTATE,
				Register: reg1,
			},
			&expr.Bitwise{
				SourceRegister: reg1,
				DestRegister:   reg1,
				Len:            4,
				Mask:           binaryutil.BigEndian.PutUint32(expr.CtStateBitESTABLISHED | expr.CtStateBitRELATED),
				Xor:            []byte{0, 0, 0, 0},
			},
			&expr.Cmp{
				Op:       expr.CmpOpNeq,
				Register: reg1,
				Data:     []byte{0, 0, 0, 0},
			},
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})

	return c.Flush()
}
