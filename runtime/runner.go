//go:build linux

package runtime

import (
	"ctrz/cgroup"
	"ctrz/logging"
	"ctrz/network"
	"ctrz/proc"
	"ctrz/spec"
	"fmt"
	"syscall"
	"time"
)

type Runtime struct {
	NetworkManager network.Manager
}

func New() Runtime {
	return Runtime{
		NetworkManager: network.Manager{
			Subnet: "10.200.1.0/24",
			Bridge: "ctrz-br0",
		},
	}
}

func (r Runtime) Run(cont *spec.ContainerSpec) error {
	containerIP, err := network.AssignContIP()
	if err != nil {
		return err
	}
	if err := network.SetupHostNetworking(); err != nil {
		return err
	}
	process := proc.Prepare(cont.Name, containerIP, cont.Command)
	if err := logging.ProcessLogs(cont.Name, process, cont.Detach); err != nil {
		return err
	}
	pid, err := proc.Run(process)
	if err != nil {
		return err
	}
	if err := cgroup.CreateAndAttach(pid, cont.CPU); err != nil {
		return err
	}
	if err := network.SetupVeth(pid); err != nil {
		return err
	}
	if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
		return err
	}
	if err := network.DenyAllElse(containerIP); err != nil {
		return err
	}
	var hostPorts []int
	var containerPorts []int
	for _, p := range cont.Ports {
		hostPort, containerPort, err := network.ExposePort(p, containerIP)
		if err != nil {
			return err
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
		PID:         pid,
		Spec:        *cont,
		StartTime:   p.Starttime,
		Started:     time.Now().Unix(),
		Cgroup:      cg,
		NetworkSpec: networkSpec,
		ProcStats:   p,
	}

	err = AttachNameToPID(container)
	if err != nil {
		return err
	}
	if !cont.Detach {
		if err := process.Wait(); err != nil {
			return err
		}
	}
	if cont.Remove && !cont.Detach {
		if err := network.RemoveContainerByName(cont.Name, false, pid, networkSpec); err != nil {
			return fmt.Errorf("Error cleaning up container: %v", err)
		}
	}
	return nil
}
