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

	lines := strings.Split(string(data), "\n")
	sockets := make(map[uint64]NetSocket)

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		inode, _ := strconv.ParseUint(fields[9], 10, 64)

		sockets[inode] = NetSocket{
			Inode:      inode,
			Proto:      proto,
			LocalAddr:  fields[1],
			RemoteAddr: fields[2],
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