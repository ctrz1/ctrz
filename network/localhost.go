package network

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

func Userland(proto, containerIP string, hostPort, containerPort int) (err error) {
	switch proto {
	// TODO: accept other protocols
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

			go func(c net.Conn) {
				if err := proxy(c, containerIP, containerPort, "tcp4"); err != nil {
					log.Printf("proxy failed: %v", err)
				}
			}(c)
		}
	default:
		return fmt.Errorf("Unsupported protocol: %s\n", proto)
	}
}

func proxy(c net.Conn, containerIP string, containerPort int, proto string) (err error) {
	defer func() {
		err = errors.Join(err, c.Close())
	}()

	dst, err := net.Dial(proto, fmt.Sprintf("[%s]:%d", containerIP, containerPort))
	if err != nil {
		log.Println("dial failed:", err)
		return err
	}
	defer func() {
		err = errors.Join(err, dst.Close())
	}()

	done := make(chan struct{}, 2)
	errCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(dst, c)
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(c, dst)
		errCh <- err
	}()

	err = errors.Join(<-errCh, <-errCh)

	<-done
	<-done
	return err
}
