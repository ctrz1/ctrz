//go:build linux

package runtime

import (
	"ctrz/cgroup"
	"ctrz/logging"
	"ctrz/network"
	"ctrz/proc"
	"ctrz/spec"
	"fmt"
	"time"
)

type Runtime struct {
	NetworkManager network.Manager
	CgroupManager  cgroup.Manager
}

func New() Runtime {
	return Runtime{
		NetworkManager: network.New(),
		CgroupManager: cgroup.New(),
	}
}

func (r Runtime) Run(cont *spec.ContainerSpec) error {
	containerIP, err := r.NetworkManager.Initialise()
	if err != nil {
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
	if err := r.CgroupManager.CreateAndAttach(pid, cont.CPU); err != nil {
		return err
	}
	networkSpec, err := r.NetworkManager.SetUp(pid, containerIP, cont.Ports)
	if err != nil {
		return err
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
