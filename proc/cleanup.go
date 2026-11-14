package proc

import (
	"ctrz/spec"
	"ctrz/utils"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func Kill(rm *spec.Removal, pid int, starttime uint64) error {
	if IsProcActive(pid, starttime) {
		fmt.Printf("Killing process %d\n", pid)
		if rm.Force {
			if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
				return fmt.Errorf("Error killing container: %v\n", err)
			}
		} else {
			if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
				return fmt.Errorf("Error killing container: %v\n", err)
			}
		}
	}
	return nil
}

func CleanBackground(proc string) error {
	utils.EnsureRoot()
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, proc)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	s := strings.Split(string(data), " ")
	pid, err := strconv.Atoi(s[0])
	if err != nil {
		return err
	}
	starttime, err := strconv.ParseUint(s[1], 10, 64)
	if err != nil {
		return err
	}
	if IsProcActive(pid, starttime) {
		out, err := exec.Command("kill", strconv.Itoa(pid)).CombinedOutput()
		if err != nil {
			log.Fatalf("%v: %s", err, out)
		}
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}
