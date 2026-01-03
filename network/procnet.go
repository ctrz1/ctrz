package network

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseProcNet(path, proto string, pid int) (map[uint64]NetSocket, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ipV6 := false
	if proto == "tcp6" || proto == "udp6" {
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

		bytesReceived, bytesSent, err := ParseProcPIDDev(pid)

		sockets[inode] = NetSocket{
			Inode:         inode,
			Proto:         proto,
			LocalAddr:     localAddr,
			RemoteAddr:    remoteAddr,
			State:         tcpState(fields[3]),
			SentBytes:     bytesSent,
			ReceivedBytes: bytesReceived,
		}
	}

	return sockets, nil
}

func ParseProcPIDDev(pid int) (bytesReceived string, bytesSent string, err error) {
	path := fmt.Sprintf("/proc/%d/net/dev", pid)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 17 || fields[0] != "eth0:" {
			continue
		}
		return fields[1], fields[9], nil
	}
	return "", "", fmt.Errorf("error parsing traffic")
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
