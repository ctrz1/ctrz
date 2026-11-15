//go:build linux
// +build linux

package proc

import (
	"ctrz/cgroup"
	"ctrz/spec"
	"ctrz/utils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

func Save(proc spec.ProcStats) (err error) {
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

func StartBackground(proc string) (int, error) {
	utils.EnsureRoot()
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return 0, err
	}

	path := filepath.Join(dir, proc)

	data, err := os.ReadFile(path)
	if err == nil {
		s := strings.Split(string(data), " ")
		pid, err := strconv.Atoi(s[0])
		if err != nil {
			return 0, err
		}
		starttime, err := strconv.ParseUint(s[1], 10, 64)
		if err != nil {
			return 0, err
		}
		if IsProcActive(pid, starttime) {
			return 0, fmt.Errorf("%s already running with PID: %d\n", proc, pid)
		}
	}

	d := exec.Command("/proc/self/exe", fmt.Sprintf("__%s", proc))

	if err := d.Start(); err != nil {
		return 0, err
	}
	cgroupManager := cgroup.New()
	if err := cgroupManager.CreateAndAttach(&cgroup.Conf{CPU: "100000 100000"}, d.Process.Pid); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}
	stat, err := ProcessStats(d.Process.Pid)
	if err != nil {
		return 0, err
	}
	daemonInfo := fmt.Sprintf("%d %d", d.Process.Pid, stat.Starttime)
	if err := os.WriteFile(path, []byte(daemonInfo), 0o755); err != nil {
		return 0, err
	}
	return d.Process.Pid, nil
}
