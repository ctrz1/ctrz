//go:build linux
// +build linux

package network

import (
	"ctrz/cgroup"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func CreateNetNs(command []string, detach bool, maxCpu string) (int, error) {
	args := append([]string{"__ctrz_init"}, command...)

	cmd := exec.Command("/proc/self/exe", args...)

	if !detach {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNET,
	}

	err := cmd.Start()
	if err != nil {
		return 0, err
	}

	pid := cmd.Process.Pid
	fmt.Printf("Started process with PID %d\n", pid)

	err = cgroup.CreateAndAttach(pid, maxCpu)
	if err != nil {
		return pid, err
	}
	if !detach {
		return pid, cmd.Wait()
	}
	return pid, nil
}

