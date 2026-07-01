//go:build linux
// +build linux

package network

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ResolveSockets(pid int) ([]NetSocket, error) {
	fds, err := SocketFDs(pid)
	if err != nil {
		return nil, err
	}

	tcp, _ := ParseProcNet("tcp", pid)
	tcp6, _ := ParseProcNet("tcp6", pid)
	udp, _ := ParseProcNet("udp", pid)
	udp6, _ := ParseProcNet("udp6", pid)

	index := map[uint64]NetSocket{}
	for _, m := range []map[uint64]NetSocket{tcp, tcp6, udp, udp6} {
		for k, v := range m {
			index[k] = v
		}
	}

	var out []NetSocket
	for _, fd := range fds {
		if s, ok := index[fd.Inode]; ok {
			out = append(out, s)
		} else {
			out = append(out, NetSocket{
				Inode: fd.Inode,
				Proto: "unknown",
				State: "LISTEN/RAW",
			})
		}
	}

	return out, nil
}


func SocketFDs(pid int) ([]SocketFD, error) {
	dir := fmt.Sprintf("/proc/%d/fd", pid)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var out []SocketFD

	for _, e := range entries {
		link, err := os.Readlink(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}

		if strings.HasPrefix(link, "socket:[") {
			inodeStr := strings.TrimSuffix(strings.TrimPrefix(link, "socket:["), "]")
			inode, _ := strconv.ParseUint(inodeStr, 10, 64)
			fd, _ := strconv.Atoi(e.Name())

			out = append(out, SocketFD{
				FD:    fd,
				Inode: inode,
			})
		}
	}

	return out, nil
}
