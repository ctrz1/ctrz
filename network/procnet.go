package network

import (
	"os"
	"strconv"
	"strings"
)

func ParseProcNet(path, proto string) (map[uint64]NetSocket, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ipV6 := false
	if (proto == "tcp6" || proto == "udp6") {
		ipV6 = true
	}

	lines := strings.Split(string(data), "\n")
	sockets := make(map[uint64]NetSocket)

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		inode, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			inode = 0
		}

		remoteAddr, err := ParseProcNetAddr(fields[1], ipV6)
		if err != nil {
			remoteAddr = fields[1]
		}
		localAddr, err := ParseProcNetAddr(fields[2], ipV6)
		if err != nil {
			localAddr = fields[2]
		}

		sockets[inode] = NetSocket{
			Inode:      inode,
			Proto:      proto,
			LocalAddr:  localAddr,
			RemoteAddr: remoteAddr,
			State:      tcpState(fields[3]),
		}
	}

	return sockets, nil
}

func tcpState(hex string) string {
	switch hex {
	case "0A":
		return "LISTEN"
	case "01":
		return "ESTABLISHED"
	case "06":
		return "TIME_WAIT"
	default:
		return hex
	}
}