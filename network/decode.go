package network

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func ParseProcNetAddr(addr string, ipv6 bool) (string, error) {
	ipHex, portHex, ok := strings.Cut(addr, ":")
	if !ok {
		return "", fmt.Errorf("invalid address format")
	}

	var ip net.IP
	var err error

	if ipv6 {
		ip, err = parseIPv6(ipHex)
	} else {
		ip, err = parseIPv4(ipHex)
	}
	if err != nil {
		return "", err
	}

	port, err := parsePort(portHex)
	if err != nil {
		return "", err
	}
	
	if ipv6 {
		return fmt.Sprintf("[%s]:%d", ip.String(), port), nil
	} else {
		return fmt.Sprintf("%s:%d", ip.String(), port), nil
	}
}

func parsePort(hexPort string) (uint16, error) {
	p, err := strconv.ParseUint(hexPort, 16, 16)
	return uint16(p), err
}

func parseIPv6(hexIP string) (net.IP, error) {
	if len(hexIP) != 32 {
		return nil, fmt.Errorf("invalid IPv6 length")
	}

	b, err := hex.DecodeString(hexIP)
	if err != nil {
		return nil, err
	}

	return net.IP(b), nil
}

func parseIPv4(hexIP string) (net.IP, error) {
	if len(hexIP) != 8 {
		return nil, fmt.Errorf("invalid IPv4 length")
	}

	b, err := hex.DecodeString(hexIP)
	if err != nil {
		return nil, err
	}

	// IPv4 is little-endian in /proc/net/*
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return net.IP(b), nil
}
