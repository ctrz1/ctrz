//go:build linux

package network

import (
	"ctrz/spec"
	"fmt"
	"log/slog"

	"github.com/google/nftables"
)

func (m Manager) removeContainerNetworking(network spec.Network) error {

	c := m.Nftables.Conn
	table := m.Nftables.Table

	var rules []*nftables.Rule

	if m.Nftables.Prerouting != nil {
		prerouting, err := c.GetRules(table, m.Nftables.Prerouting)
		if err != nil {
			return err
		}
		rules = append(rules, prerouting...)
	}

	if m.Nftables.Output != nil {
		output, err := c.GetRules(table, m.Nftables.Output)
		if err != nil {
			return err
		}
		rules = append(rules, output...)
	}

	if m.Nftables.Forward != nil {
		forward, err := c.GetRules(table, m.Nftables.Forward)
		if err != nil {
			return err
		}
		rules = append(rules, forward...)
	}

	if len(rules) == 0 {
		return nil
	}

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
