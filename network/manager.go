//go:build linux

package network

import (
	"ctrz/spec"
	"fmt"
	"log"
	"net"
	"syscall"
)

type Manager struct {
	Subnet             string
	Gateway            string
	Bridge             string
	ContainerInterface string
}

func New(subnet, bridge, containerInterface string) (Manager, error) {
	gateway, err := gateway(subnet)
	if err != nil {
		return Manager{}, fmt.Errorf("Error creating network manager: %v\n", err)
	}
	return Manager{
		Subnet:             subnet,             //"10.200.1.0/24",
		Gateway:            gateway,            //"10.200.1.1/24",
		Bridge:             bridge,             //"ctrz-br0",
		ContainerInterface: containerInterface, //"ctrz0",
	}, nil
}

func gateway(subnet string) (string, error) {
	_, base, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", fmt.Errorf("Could not parse subnet (%s): %v\n", subnet, err)
	}
	gateway := base.IP
	gateway[3]++
	return fmt.Sprintf("%s/24", gateway.String()), nil
}

func (m Manager) Initialise() (string, error) {
	return AssignContIP()
}

func (m Manager) SetUp(pid int, ip string, ports []string) (spec.Network, error) {
	if err := m.SetupHostNetworking(); err != nil {
		return spec.Network{}, err
	}
	if err := m.SetupVeth(pid); err != nil {
		return spec.Network{}, err
	}
	if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
		return spec.Network{}, err
	}
	if err := DenyAllElse(ip); err != nil {
		return spec.Network{}, err
	}
	var hostPorts []int
	var containerPorts []int
	for _, p := range ports {
		hostPort, containerPort, err := ExposePort(p, ip)
		if err != nil {
			return spec.Network{}, err
		}

		hostPorts = append(hostPorts, hostPort)
		containerPorts = append(containerPorts, containerPort)
	}

	networkSpec := spec.Network{
		IP: ip,
	}

	for i, containerPort := range containerPorts {
		networkSpec.Ports = append(networkSpec.Ports, spec.PortMapping{
			ContainerPort: containerPort,
			HostPort:      hostPorts[i],
		})
	}

	return networkSpec, nil
}

func (m Manager) Configure() {

}

func (m Manager) Cleanup(network spec.Network) error {
	if err := removeIPTableRules(network); err != nil {
		log.Fatal(err)
	}
	if err := RemoveContIP(network.IP); err != nil {
		return err
	}
	return nil
}
