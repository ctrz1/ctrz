//go:build linux
// +build linux

package proc

import (
	"ctrz/spec"
	"ctrz/utils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func Prepare(name string, command []string) *exec.Cmd {
	args := append([]string{"__ctrz_init"}, command...)

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

func Save(proc spec.ProcStats) error {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return err
	}
	pidfile := filepath.Join(dir, "pids")
	_, err = os.Stat(pidfile)
	var f *os.File
	if err == nil {
		f, err = os.OpenFile(pidfile, os.O_WRONLY|os.O_APPEND, 0o755)
		if err != nil {
			return fmt.Errorf("Error opening PID file file: %v\n", err)
		}
	} else {
		f, err = os.OpenFile(pidfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o755)
		if err != nil {
			return fmt.Errorf("Error opening PID file file: %v\n", err)
		}
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()
	if _, err := fmt.Fprintf(f, "%d %d\n", proc.PID, proc.Starttime); err != nil {
		return fmt.Errorf("Error adding process to PID file: %v\n", err)
	}
	return nil
}
