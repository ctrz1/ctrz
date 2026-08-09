//go:build linux

package network

import (
	"ctrz/spec"
	"syscall"
)

type Manager struct {
	Subnet             string
	Bridge             string
	ContainerInterface string
}

func New() Manager {
	return Manager{
		Subnet:             "10.200.1.0/24",
		Bridge:             "ctrz-br0",
		ContainerInterface: "ctrz0",
	}
}

func (m Manager) Initialise() (string, error) {
	return AssignContIP()
}

func (m Manager) SetUp(pid int, ip string, ports []string) (spec.Network, error) {
	if err := SetupHostNetworking(); err != nil {
		return spec.Network{}, err
	}
	if err := SetupVeth(pid); err != nil {
		return spec.Network{}, nil
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

func (m Manager) Cleanup() {

}
