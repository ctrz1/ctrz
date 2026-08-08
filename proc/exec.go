//go:build linux
// +build linux

package proc

import (
	"fmt"
	"os/exec"
	"syscall"
)

func Prepare(name, ip string, command []string) *exec.Cmd {
	args := append([]string{name}, command...)
	args = append([]string{ip}, args...)
	args = append([]string{"__ctrz_init"}, args...)

	process := exec.Command("/proc/self/exe", args...)

	process.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:
		// syscall.CLONE_NEWUSER | --> can be used for rootless containers later. Ignore for now
		syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID,

		Unshareflags:               syscall.CLONE_NEWNS,
		GidMappingsEnableSetgroups: false,
	}
	return process
}

func Run(process *exec.Cmd) (int, error) {
	if err := process.Start(); err != nil {
		fmt.Printf("Error starting process: %v\n", err)
		return 0, err
	}

	return process.Process.Pid, nil
}
