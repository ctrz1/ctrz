//go:build linux

package network

import (
	"fmt"

	"github.com/google/nftables"
)

func (m *Manager) initialiseNftables() error {
	c, err := nftables.New()
	if err != nil {
		return fmt.Errorf("Error initialising nftable connection: %v\n", err)
	}

	m.Nftables.Conn = c

	table := c.CreateTable(&nftables.Table{
		Name:   "ctrz",
		Family: nftables.TableFamilyIPv4,
	})

	m.Nftables.Table = table

	m.Nftables.Prerouting = c.AddChain(&nftables.Chain{
		Table:    table,
		Name:     "prerouting",
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityNATDest,
	})

	m.Nftables.Output = c.AddChain(&nftables.Chain{
		Table:    table,
		Name:     "output",
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityNATDest,
	})

	m.Nftables.Forward = c.AddChain(&nftables.Chain{
		Table:    table,
		Name:     "forward",
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
	})

	if m.Nftables.Conn == nil || m.Nftables.Table == nil {
		fmt.Println("connection and table were not set properly")
	}

	return nil
}
