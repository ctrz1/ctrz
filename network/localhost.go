//go:build linux
// +build linux

package network

import (
	"fmt"
	"io"
	"log"
	"net"
)

func Userland(proto string, containerIP string, hostPort int, containerPort int) error {
	switch proto {
	//TODO: accept other protocols
	case "tcp":
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", hostPort))
		if err != nil {
			return err
		}
		for {
		    c, err := ln.Accept()
			if err != nil {
				log.Printf("accept failed: %v\n", err)
				continue
			}
		    go proxy(c, containerIP, containerPort, "tcp4")
		}
	default:
		return fmt.Errorf("Unsupported protocol: %s\n", proto)
	}
	
}

func proxy(c net.Conn, containerIP string, containerPort int, proto string) {
	defer c.Close()

	dst, err := net.Dial(proto, fmt.Sprintf("[%s]:%d", containerIP, containerPort))
	if err != nil {
		log.Println("dial failed:", err)
		return
	}
	defer dst.Close()

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(dst, c)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(c, dst)
		done <- struct{}{}
	}()

	<-done
	<-done
}
