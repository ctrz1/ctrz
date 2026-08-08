//go:build linux
// +build linux

package network

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"ctrz/spec"
)

func SetupHostNetworking() error {
	// Enable IPv4 forwarding
	if out, err := exec.Command(
		"sysctl",
		"-w",
		"net.ipv4.ip_forward=1",
	).CombinedOutput(); err != nil {
		return fmt.Errorf("enable ip_forward failed: %v: %s", err, out)
	}

	if out, err := exec.Command(
		"sysctl",
		"-w",
		"net.ipv4.conf.all.route_localnet=1",
	).CombinedOutput(); err != nil {
		return fmt.Errorf("enable ip_forward failed: %v: %s", err, out)
	}

	// Create bridge if missing
	if err := exec.Command("ip", "link", "show", "ctrz-br0").Run(); err != nil {
		if out, err := exec.Command("ip", "link", "add", "ctrz-br0", "type", "bridge").CombinedOutput(); err != nil {
			return fmt.Errorf("create bridge failed: %v: %s", err, out)
		}
		if out, err := exec.Command("ip", "addr", "add", "10.200.1.1/24", "dev", "ctrz-br0").CombinedOutput(); err != nil {
			return fmt.Errorf("assign bridge ip failed: %v: %s", err, out)
		}
	}

	if out, err := exec.Command("ip", "link", "set", "ctrz-br0", "up").CombinedOutput(); err != nil {
		return fmt.Errorf("bring bridge up failed: %v: %s", err, out)
	}

	// NAT
	if err := ensureRule(
		[]string{
			"-t", "nat",
			"-C", "POSTROUTING",
			"-s", "10.200.1.0/24",
			"-j", "MASQUERADE",
		},
		[]string{
			"-t", "nat",
			"-A", "POSTROUTING",
			"-s", "10.200.1.0/24",
			"-j", "MASQUERADE",
		},
	); err != nil {
		return err
	}
	if err := ensureRule(
		[]string{
			"-t", "nat",
			"-A", "POSTROUTING",
			"-s", "127.0.0.0/8",
			"-d", "10.200.1.0/24",
			"-j", "MASQUERADE",
		},
		[]string{
			"-t", "nat",
			"-A", "POSTROUTING",
			"-s", "127.0.0.0/8",
			"-d", "10.200.1.0/24",
			"-j", "MASQUERADE",
		},
	); err != nil {
		return err
	}

	// Allow outbound forwarding
	if err := ensureRule(
		[]string{
			"-C", "FORWARD",
			"-i", "ctrz-br0",
			"-j", "ACCEPT",
		},
		[]string{
			"-A", "FORWARD",
			"-i", "ctrz-br0",
			"-j", "ACCEPT",
		},
	); err != nil {
		return err
	}

	// Allow return traffic
	if err := ensureRule(
		[]string{
			"-C", "FORWARD",
			"-o", "ctrz-br0",
			"-m", "conntrack",
			"--ctstate", "ESTABLISHED,RELATED",
			"-j", "ACCEPT",
		},
		[]string{
			"-A", "FORWARD",
			"-o", "ctrz-br0",
			"-m", "conntrack",
			"--ctstate", "ESTABLISHED,RELATED",
			"-j", "ACCEPT",
		},
	); err != nil {
		return err
	}

	return nil
}

func SetupNetns(containerIP string) error {
	if out, err := exec.Command("ip", "link", "set", "lo", "up").CombinedOutput(); err != nil {
		return fmt.Errorf("loopback up failed: %v: %s", err, out)
	}
	if out, err := exec.Command("ip", "addr", "add", containerIP+"/24", "dev", "ctrz0").CombinedOutput(); err != nil {
		return fmt.Errorf("assign container ip failed: %v: %s", err, out)
	}
	if out, err := exec.Command("ip", "link", "set", "ctrz0", "up").CombinedOutput(); err != nil {
		return fmt.Errorf("container interface up failed: %v: %s", err, out)
	}
	if out, err := exec.Command("ip", "route", "add", "default", "via", "10.200.1.1").CombinedOutput(); err != nil {
		return fmt.Errorf("default route failed: %v: %s", err, out)
	}

	return nil
}

func ExposePort(ports, containerIP string) (int, int, error) {
	pm, err := parsePorts(ports)
	if err != nil {
		return -1, -1, err
	}

	cmds := [][]string{
		{
			"iptables",
			"-t", "nat",
			"-I", "PREROUTING", "1",
			"-p", "tcp",
			"--dport", strconv.Itoa(pm.HostPort),
			"-j", "DNAT",
			"--to-destination",
			fmt.Sprintf("%s:%d", containerIP, pm.ContainerPort),
		},
		{
			"iptables",
			"-I", "FORWARD", "1",
			"-p", "tcp",
			"-d", containerIP,
			"--dport", strconv.Itoa(pm.ContainerPort),
			"-j", "ACCEPT",
		},
		{
			"iptables",
			"-t", "nat",
			"-I", "OUTPUT", "1",
			"-p", "tcp",
			"-d", "127.0.0.0/8",
			"--dport", strconv.Itoa(pm.HostPort),
			"-j", "DNAT", "--to-destination", fmt.Sprintf("%s:%d", containerIP, pm.ContainerPort),
		},
	}

	for _, c := range cmds {
		out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
		if err != nil {
			return -1, -1, fmt.Errorf("%v: %s", err, out)
		}
	}

	return pm.HostPort, pm.ContainerPort, nil
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

func ruleExists(args ...string) (bool, error) {
	cmd := exec.Command("iptables", args...)
	err := cmd.Run()

	if err == nil {
		return true, nil
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return false, err
	}

	if exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, err
}

func ensureRule(checkArgs, addArgs []string) error {
	exists, err := ruleExists(checkArgs...)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	out, err := exec.Command("iptables", addArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("iptables failed: %v: %s", err, out)
	}

	return nil
}
