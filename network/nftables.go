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

	table := c.AddTable(&nftables.Table{
		Name:   "ctrz",
		Family: nftables.TableFamilyIPv4,
	})

	m.Nftables.Table = table

	if err := c.Flush(); err != nil {
		return fmt.Errorf("Error adding ctrz table: %v\n", err)
	}

	return nil
}

func (m *Manager) addNftChains() error {
	c := m.Nftables.Conn
	table := m.Nftables.Table

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

	if err := c.Flush(); err != nil {
		return fmt.Errorf("Error creating ctrz chains: %v\n", err)
	}
	return nil
}

func (m *Manager) getChains() error {
	c := m.Nftables.Conn

	chains, err := c.ListChains()
	if err != nil {
		return fmt.Errorf("Error getting chains: %v\n", err)
	}

	for _, chain := range chains {
			switch chain.Name {
			case "prerouting":
				m.Nftables.Prerouting = chain
			case "output":
				m.Nftables.Output = chain
			case "forward":
				m.Nftables.Forward = chain
			default:
				continue
			}
	}
	return nil
}
