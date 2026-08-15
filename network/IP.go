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
		return []string{}, fmt.Errorf("Error reading IP file: %v\n", err)
	}
	data := string(b)
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
	if err := os.Remove(filepath.Join(dir, "containers", "IP")); err != nil {
		fmt.Printf("Warning: Could not find IP file to remove container IP: %v\n", err)
		return nil
	}
	f, err := os.OpenFile(filepath.Join(dir, "containers", "IP"), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o755)
	if err != nil {
		return fmt.Errorf("Error opening IP file: %v\n", err)
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()
	for _, ip := range remainingIPs {
		// If this fails mid-loop, it leaves data inconsistent with existing containers
		if _, err := fmt.Fprintf(f, "%s\n", ip); err != nil {
			return err
		}
	}
	return nil
}
