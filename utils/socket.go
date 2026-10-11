package utils

import (
	"ctrz/spec"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func ListenToSocket() (spec.ContainerSocket, error) {
	var cont spec.ContainerSocket
	baseDir, err := CtrzStateDir()
	if err != nil {
		return cont, err
	}

	if err := os.Remove(filepath.Join(baseDir, "ctrz.sock")); err != nil {
		fmt.Println(err)
	}

	buf := make([]byte, 512)

	l, err := net.Listen("unix", filepath.Join(baseDir, "ctrz.sock"))
	if err != nil {
		return cont, err
	}
	c, err := l.Accept()
	if err != nil {
		return cont, err
	}
	defer c.Close()

	nr, err := c.Read(buf)
	if err != nil {
		return cont, err
	}

	data := buf[0:nr]
	if err := json.Unmarshal(data, &cont); err != nil {
		return cont, err
	}
	fmt.Printf("Received data:\n\n%+v\n", cont)
	return cont, nil
}

func WriteToSocket(cont spec.ContainerSocket) error {
	baseDir, err := CtrzStateDir()
	if err != nil {
		return err
	}

	c, err := net.Dial("unix", filepath.Join(baseDir, "ctrz.sock"))
	if err != nil {
		return err
	}
	defer c.Close()

	b, err := json.Marshal(cont)
	if err != nil {
		return err
	}
	_, err = c.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// func ensureSocket() error {
// 	baseDir, err := CtrzStateDir()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
