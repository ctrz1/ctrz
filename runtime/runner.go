//go:build linux

package runtime

import (
	"ctrz/cgroup"
	"ctrz/logging"
	"ctrz/network"
	"ctrz/proc"
	"ctrz/spec"
	"ctrz/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Runtime struct {
	NetworkManager network.Manager
	CgroupManager  cgroup.Manager
}

func New() Runtime {
	return Runtime{
		NetworkManager: network.New("10.200.1.0/24", "ctrz-br0", "ctrz0"),
		CgroupManager:  cgroup.New(),
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

	cgPath, err := r.CgroupManager.Path(pid)
	if err != nil {
		return err
	}

	container := spec.Container{
		PID:         pid,
		Spec:        *cont,
		StartTime:   p.Starttime,
		Started:     time.Now().Unix(),
		Cgroup:      cgPath,
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
		r.Remove(&spec.Removal{
			Name:     cont.Name,
			Force:    false,
			All:      false,
			Inactive: false,
		})
	}
	return nil
}

func (r Runtime) Container(name string) (*spec.Container, error) {
	cont, err := GetContainerDataFromName(name)
	return &cont, err
}

func (r Runtime) Remove(rm *spec.Removal) error {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return err
	}

	var containers []string

	if rm.Inactive || rm.All {
		containers, err = RetrieveAllContainers()
		if err != nil {
			return fmt.Errorf("Could not retrieve list of containers: %v\n", err)
		}
	} else {
		containers = append(containers, rm.Name)
	}

	for _, c := range containers {
		cont, err := GetContainerDataFromName(c)
		if err != nil {
			return fmt.Errorf("Error removing container: %v\n", err)
		}
		if err := proc.Kill(rm, cont.PID, cont.StartTime); err != nil {
			return fmt.Errorf("Error cleaning up process: %v\n", err)
		}
		if err := r.NetworkManager.Cleanup(cont.NetworkSpec); err != nil {
			return fmt.Errorf("Error cleaning up container networking: %v\n", err)
		}
		if err := os.RemoveAll(filepath.Join(dir, "containers", c)); err != nil {
			return fmt.Errorf("Error deleting container data: %v\n", err)
		}
	}
	return nil
}
