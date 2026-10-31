//go:build linux

package runtime

import (
	"context"
	"ctrz/cgroup"
	"ctrz/fs"
	"ctrz/logging"
	"ctrz/network"
	"ctrz/proc"
	"ctrz/spec"
	"ctrz/utils"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type Runtime struct {
	NetworkManager    *network.Manager
	CgroupManager     cgroup.Manager
	FileSystemManager fs.Manager
}

func New() (Runtime, error) {
	netManager, err := network.New("10.200.1.0/24", "ctrz-br0", "ctrz0")
	if err != nil {
		return Runtime{}, err
	}
	return Runtime{
		NetworkManager:    &netManager,
		CgroupManager:     cgroup.New(),
		FileSystemManager: fs.New(),
	}, nil
}

func (r Runtime) Run(cont *spec.ContainerSpec) error {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return err
	}
	sockPath := filepath.Join(dir, "ctrz.sock")
	containerIP, err := r.NetworkManager.Initialise()
	if err != nil {
		return err
	}
	process := proc.Prepare(cont.Name, cont.Command)
	if err := logging.ProcessLogs(cont.Name, process, cont.Detach); err != nil {
		return err
	}
	pid, err := proc.Run(process)
	if err != nil {
		return err
	}
	p, err := proc.ProcessStats(pid)
	if err != nil {
		return err
	}
	if err := proc.Save(p); err != nil {
		return fmt.Errorf("Error saving process: %v\n", err)
	}
	if err := r.CgroupManager.CreateAndAttach(&cgroup.Conf{
		CPU:    cont.CPU,
		Memory: cont.MemoryLimit,
	}, pid); err != nil {
		return err
	}
	networkSpec, err := r.NetworkManager.SetUp(pid, containerIP, cont.Ports)
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := utils.WriteToSocket(sctx, sockPath, spec.ContainerSocket{
		IP:          containerIP,
		ID:          cont.Name,
		Env:         cont.Environment,
		ExposePorts: len(networkSpec.Ports) > 0,
	}); err != nil {
		fmt.Printf("Error writing to socket: %v\n", err)
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
		if err := r.Remove(&spec.Removal{
			Name:     cont.Name,
			Force:    false,
			All:      false,
			Inactive: false,
		}); err != nil {
			return err
		}
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
		all, err := RetrieveAllContainers()
		if err != nil {
			return fmt.Errorf("Could not retrieve list of containers: %v\n", err)
		}
		for _, c := range all {
			cont, err := GetContainerDataFromName(c)
			if err != nil {
				continue
			}
			if !proc.IsProcActive(cont.PID, cont.StartTime) || rm.All {
				containers = append(containers, c)
			}
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

func (r Runtime) Clean() error {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return err
	}
	containers := getContainerFromPID(dir)
	rm := &spec.Removal{
		Force: true,
	}
	for _, cont := range containers {
		if err := proc.Kill(rm, cont.pid, cont.starttime); err != nil {
			continue
		}
	}
	if err := r.NetworkManager.Clean(); err != nil {
		return err
	}
	if err := fs.Clean(); err != nil {
		return err
	}
	return nil
}
