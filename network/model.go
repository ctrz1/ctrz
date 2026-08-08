package network

import (
	"net"
)

type TCPSocket struct {
	LocalIP    net.IP
	LocalPort  uint16
	RemoteIP   net.IP
	RemotePort uint16
	State      string
	Inode      uint64
}

type SocketFD struct {
	FD    int
	Inode uint64
}

type NetSocket struct {
	Inode         uint64
	Proto         string // tcp, tcp6, udp, udp6
	LocalAddr     string
	RemoteAddr    string
	State         string
	SentBytes     string
	ReceivedBytes string
}

type Interface struct {
	Name          string
	SentBytes     uint64
	ReceivedBytes uint64
}

type VethPair struct {
	ContainerInterface string
	HostInterface      string
}
