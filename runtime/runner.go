//go:build linux

package runtime

import (
	"ctrz/cgroup"
	"ctrz/network"
	"ctrz/proc"
	"ctrz/spec"
	"fmt"
	"log"
	"strings"
	"syscall"
	"time"
)

func New(name, cpu string, portMappings []string, remove, detach bool, args []string) (*spec.ContainerSpec, error) {
	if name == "" {
		name = GenerateRandomContName()
		for !CheckContName(name) {
			name = GenerateRandomContName()
		}
	} else {
		if !CheckContName(name) {
			return nil, fmt.Errorf("Container '%s' already exists. Either choose a different name or remove the existing container", name)
		}
	}

	var command string
	for _, v := range args {
		command += fmt.Sprintf("%s ", v)
	}
	command = strings.Trim(command, " ")

	spec := spec.ContainerSpec{
		Name:    name,
		CPU:     cpu,
		Command: args,
		Remove:  remove,
		Detach:  detach,
		Ports:   portMappings,
	}
	return &spec, nil
}

func Run(cont *spec.ContainerSpec) error {
	containerIP, err := network.AssignContIP()
	if err != nil {
		log.Fatal(err)
	}
	if err := network.SetupHostNetworking(); err != nil {
		log.Fatal(err)
	}
	pid, process, err := network.CreateNetNs(cont.Command, cont.CPU, cont.Name, containerIP, cont.Detach)
	if err != nil {
		log.Fatal(err)
	}
	if err := network.SetupVeth(pid); err != nil {
		log.Fatal(err)
	}
	if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
		log.Fatal(err)
	}
	if err := network.DenyAllElse(containerIP); err != nil {
		log.Fatal(err)
	}
	var hostPorts []int
	var containerPorts []int
	for _, p := range cont.Ports {
		hostPort, containerPort, err := network.ExposePort(p, containerIP)
		if err != nil {
			log.Fatal(err)
		}

		hostPorts = append(hostPorts, hostPort)
		containerPorts = append(containerPorts, containerPort)
	}

	networkSpec := spec.Network{
		IP: containerIP,
	}

	for i, containerPort := range containerPorts {
		networkSpec.Ports = append(networkSpec.Ports, spec.PortMapping{
			ContainerPort: containerPort,
			HostPort:      hostPorts[i],
		})
	}

	p, err := proc.ProcessStats(pid)
	if err != nil {
		return err
	}

	cg, err := cgroup.PathForPID(pid)

	container := spec.Container{
		PID: pid,
		Spec: *cont,
		StartTime: p.Starttime,
		Started: time.Now().Unix(),
		Cgroup: cg,
		NetworkSpec: networkSpec,
		ProcStats: p,
	}

	err = AttachNameToPID(container)
	if err != nil {
		log.Fatal(err)
	}
	if !cont.Detach {
		if err := process.Wait(); err != nil {
			log.Fatal(err)
		}
	}
	if cont.Remove && !cont.Detach {
		if err := network.RemoveContainerByName(cont.Name, false, pid, networkSpec); err != nil {
			log.Fatalf("Error cleaning up container: %v", err)
		}
	}
	return nil
}
