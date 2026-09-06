//go:build linux

package network

import (
	"ctrz/spec"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"

	"github.com/google/nftables"
)

const (
	IPSUBNET = "10.200.1.0/24"

	IPTABLES          = "iptables"
	PREROUTING        = "PREROUTING"
	OUTPUT            = "OUTPUT"
	INPUT             = "INPUT"
	POSTROUTING       = "POSTROUTING"
	MASQUERADE        = "MASQUERADE"
	DNAT              = "DNAT"
	FORWARD           = "FORWARD"
	ACCEPT            = "ACCEPT"
	DROP              = "DROP"
	NAT               = "nat"
	PROTOCOL          = "-p"
	TABLE             = "-t"
	DELETE            = "-D"
	TCP               = "tcp"
	JUMP              = "-j"
	SOURCE            = "-s"
	MATCH             = "-m"
	CHECK             = "-C"
	DESTINATION       = "-d"
	DESTINATION_PORT  = "--dport"
	SOURCE_PORT       = "--sport"
	CONNECTION_STATES = "--ctstate"
	CONNTRACK         = "conntrack"
	TO_DESTINATION    = "--to-destination"
)

func removeIPTableRules(network spec.Network) error {
	for i := range network.Ports {
		cmds := [][]string{
			{
				IPTABLES, TABLE, NAT, DELETE, PREROUTING,
				PROTOCOL, TCP, DESTINATION_PORT, strconv.Itoa(network.Ports[i].HostPort),
				JUMP, DNAT,
				TO_DESTINATION, fmt.Sprintf("%s:%d", network.IP, network.Ports[i].ContainerPort),
			},
			{
				IPTABLES, TABLE, NAT, DELETE, OUTPUT,
				PROTOCOL, TCP, DESTINATION_PORT, strconv.Itoa(network.Ports[i].HostPort),
				JUMP, DNAT,
				TO_DESTINATION, fmt.Sprintf("%s:%d", network.IP, network.Ports[i].ContainerPort),
			},
			{
				IPTABLES, TABLE, NAT, DELETE, POSTROUTING, SOURCE,
				IPSUBNET, JUMP, MASQUERADE,
			},
			{
				IPTABLES, DELETE, FORWARD,
				PROTOCOL, TCP, DESTINATION, network.IP,
				DESTINATION_PORT, strconv.Itoa(network.Ports[i].ContainerPort),
				JUMP, ACCEPT,
			},
			{
				IPTABLES, DELETE, FORWARD,
				PROTOCOL, TCP, SOURCE, network.IP,
				SOURCE_PORT, strconv.Itoa(network.Ports[i].ContainerPort),
				JUMP, ACCEPT,
			},
			{
				IPTABLES, DELETE, FORWARD,
				MATCH, CONNTRACK,
				CONNECTION_STATES, "ESTABLISHED,RELATED",
				JUMP, ACCEPT,
			},
			{
				IPTABLES, DELETE, INPUT, PROTOCOL, TCP,
				DESTINATION_PORT, strconv.Itoa(network.Ports[i].ContainerPort),
				JUMP, ACCEPT,
			},
			{
				IPTABLES, DELETE, FORWARD,
				DESTINATION, network.IP,
				JUMP, DROP,
			},
		}

		for _, c := range cmds {
			checkCmd := append([]string(nil), c...)
			for i, v := range checkCmd {
				if v == DELETE {
					checkCmd[i] = CHECK
					break
				}
			}
			out, err := exec.Command(checkCmd[0], checkCmd[1:]...).CombinedOutput()
			if err == nil {
				out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
				if err != nil {
					return fmt.Errorf("%v: %s", err, out)
				}
				continue
			}
			exitErr, ok := err.(*exec.ExitError)
			if !ok {
				return err
			}
			if exitErr.ExitCode() == 1 {
				continue
			}
			return fmt.Errorf("%v: %s", err, out)
		}
	}

	return nil
}

func (m Manager) removeContainerNetworking(network spec.Network) error {

	c := m.Nftables.Conn
	table := m.Nftables.Table

	var rules []*nftables.Rule

	prerouting, err := c.GetRules(table, m.Nftables.Prerouting)
	if err != nil {
		return err
	}
	rules = append(rules, prerouting...)

	output, err := c.GetRules(table, m.Nftables.Output)
	if err != nil {
		return err
	}
	rules = append(rules, output...)

	forward, err := c.GetRules(table, m.Nftables.Forward)
	if err != nil {
		return err
	}
	rules = append(rules, forward...)

	for i := range network.Ports {
		hostPort := network.Ports[i].HostPort
		containerPort := network.Ports[i].ContainerPort
		ruleID := fmt.Sprintf("ctrz:%s:%d:%d", network.IP, hostPort, containerPort)

		for _, rule := range rules {
			if string(rule.UserData) == ruleID {
				if err := c.DelRule(rule); err != nil {
					slog.Error("deleting rule", "ruleID", ruleID, "from chain", "chainName", rule.Chain.Name, ":", "error", err)
				}
			}
		}
	}
	if err := c.Flush(); err != nil {
		return fmt.Errorf("Error flushing delete rules: %v\n", err)
	}
	return nil
}
