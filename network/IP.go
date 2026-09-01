package network

import (
	"ctrz/utils"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

/**
*	This implementation of IP allocation is prone to race conditions
*	If two instances of ctrz try to create a container at the same, they might end up with the same IP
**/

func AssignContIP() (ip string, err error) {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return "", fmt.Errorf("Error retrieving ctrz state directory: %v\n", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "containers"), 0o755); err != nil {
		return "", fmt.Errorf("Error creating necessary directory: %v\n", err)
	}
	_, err = os.Stat(filepath.Join(dir, "containers", "IP"))
	if err != nil {
		newFile, err := os.OpenFile(filepath.Join(dir, "containers", "IP"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o755)
		if err != nil {
			return "", fmt.Errorf("Error creating IP file: %v\n", err)
		}
		defer func() {
			err = errors.Join(err, newFile.Close())
		}()
		containerIP := randIP([]string{})
		if err := writeIPToFile(containerIP, dir); err != nil {
			return "", err
		}
		return containerIP, nil
	}
	takenIPs, err := assignedIPs(dir)
	if err != nil {
		return "", err
	}
	containerIP := randIP(takenIPs)
	if err := writeIPToFile(containerIP, dir); err != nil {
		return "", err
	}
	return containerIP, nil
}

func assignedIPs(dir string) ([]string, error) {
	b, err := os.ReadFile(filepath.Join(dir, "containers", "IP"))
	if err != nil {
		return nil, fmt.Errorf("Error reading IP file: %v\n", err)
	}

	data := strings.TrimSpace(string(b))
	if data == "" {
		return []string{}, nil
	}

	return strings.Split(data, "\n"), nil
}

func randIP(takenIPs []string) string {
	baseIP := "10.200.1"
	client := 0
	var containerIP string
	for client < 2 {
		client = rand.Intn(256)
		containerIP = fmt.Sprintf("%s.%d", baseIP, client)
		if !slices.Contains(takenIPs, containerIP) {
			break
		}
		client = 0
	}
	return containerIP
}

func writeIPToFile(IP, dir string) (err error) {
	f, err := os.OpenFile(filepath.Join(dir, "containers", "IP"), os.O_WRONLY|os.O_APPEND, 0o755)
	if err != nil {
		return fmt.Errorf("Error opening IP file: %v\n", err)
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()
	if _, err := fmt.Fprintf(f, "%s\n", IP); err != nil {
		return fmt.Errorf("Error adding container IP to IP file: %v\n", err)
	}
	return nil
}

func RemoveContIP(containerIP string) (err error) {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return fmt.Errorf("Error retrieving ctrz state directory: %v\n", err)
	}
	takenIPs, err := assignedIPs(dir)
	if err != nil {
		return err
	}
	var remainingIPs []string
	for _, ip := range takenIPs {
		if ip == containerIP {
			continue
		}
		remainingIPs = append(remainingIPs, ip)
	}
	f, err := os.OpenFile(filepath.Join(dir, "containers", "IP.temp"), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o755)
	if err != nil {
		return fmt.Errorf("Error opening IP file: %v\n", err)
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()
	for _, ip := range remainingIPs {
		if _, err := fmt.Fprintf(f, "%s\n", ip); err != nil {
			if rmErr := os.Remove(f.Name()); rmErr != nil {
				return fmt.Errorf("Error cleaning up IP.temp: %v\nError printing to IP.temp: %v\n", rmErr, err)
			}
			return fmt.Errorf("Error printing to IP.temp: %v\n", err)
		}
	}
	if err := os.Rename(f.Name(), filepath.Join(dir, "containers", "IP")); err != nil {
		if rmErr := os.Remove(f.Name()); rmErr != nil {
			return fmt.Errorf("Error cleaning up IP.temp: %v\nError renamining IP.temp to IP: %v\n", rmErr, err)
		}
		return fmt.Errorf("Error renamining IP.temp to IP: %v\n", err)
	}
	return nil
}
