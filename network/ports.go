//go:build linux

package network

import (
	"fmt"
	"net"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

const (
	reg1 = 0x1
	reg2 = 0x2
)

func (m Manager) exposePort(ports, containerIP string) (int, int, error) {
	ip := net.ParseIP(containerIP).To4()
	if ip == nil {
		return -1, -1, fmt.Errorf("invalid container IP: %s\n", containerIP)
	}

	pm, err := parsePorts(ports)
	if err != nil {
		return -1, -1, fmt.Errorf("Error parsing ports: %v\n", err)
	}

	c := m.Nftables.Conn
	table := m.Nftables.Table

	// https://github.com/kubernetes-sigs/kube-network-policies/blob/89fb3de67c61ef275d4c8b9a5d632ad99ea03cc1/pkg/dns/dnsagent.go#L172
	c.AddRule(&nftables.Rule{
		Table: table,
		Chain: m.Nftables.Prerouting,
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyL4PROTO,
				Register: reg1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     []byte{unix.IPPROTO_TCP},
			},

			// tcp dport
			&expr.Payload{
				DestRegister: reg1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2,
				Len:          2,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     binaryutil.BigEndian.PutUint16(uint16(pm.HostPort)),
			},

			// DNAT containerIP:containerPort
			&expr.Immediate{
				Register: reg1,
				Data:     ip,
			},
			&expr.Immediate{
				Register: reg2,
				Data:     binaryutil.BigEndian.PutUint16(uint16(pm.ContainerPort)),
			},
			&expr.NAT{
				Type:        expr.NATTypeDestNAT,
				Family:      unix.NFPROTO_IPV4,
				RegAddrMin:  reg1,
				RegProtoMin: reg2,
			},
		},
	})

	c.AddRule(&nftables.Rule{
		Table: table,
		Chain: m.Nftables.Output,
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyL4PROTO,
				Register: reg1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     []byte{unix.IPPROTO_TCP},
			},

			// ip daddr 127.0.0.0/8
			&expr.Payload{
				DestRegister: reg1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       16,
				Len:          4,
			},
			&expr.Bitwise{
				SourceRegister: reg1,
				DestRegister:   reg1,
				Len:            4,
				Mask:           []byte{255, 0, 0, 0},
				Xor:            []byte{0, 0, 0, 0},
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     net.IPv4(127, 0, 0, 0).To4(),
			},

			// tcp dport HOST_PORT
			&expr.Payload{
				DestRegister: reg1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2,
				Len:          2,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     binaryutil.BigEndian.PutUint16(uint16(pm.HostPort)),
			},

			// DNAT
			&expr.Immediate{
				Register: reg1,
				Data:     ip,
			},
			&expr.Immediate{
				Register: reg2,
				Data:     binaryutil.BigEndian.PutUint16(uint16(pm.ContainerPort)),
			},
			&expr.NAT{
				Type:        expr.NATTypeDestNAT,
				Family:      unix.NFPROTO_IPV4,
				RegAddrMin:  reg1,
				RegProtoMin: reg2,
			},
		},
	})

	c.AddRule(&nftables.Rule{
		Table: table,
		Chain: m.Nftables.Forward,
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyL4PROTO,
				Register: reg1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     []byte{unix.IPPROTO_TCP},
			},

			// ip daddr containerIP
			&expr.Payload{
				DestRegister: reg1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       16,
				Len:          4,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     ip,
			},

			// tcp dport containerPort
			&expr.Payload{
				DestRegister: reg1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2,
				Len:          2,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: reg1,
				Data:     binaryutil.BigEndian.PutUint16(uint16(pm.ContainerPort)),
			},

			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})

	if err := c.Flush(); err != nil {
		return -1, -1, fmt.Errorf("Error flushing rules: %v\n", err)
	}

	return pm.HostPort, pm.ContainerPort, nil
}
