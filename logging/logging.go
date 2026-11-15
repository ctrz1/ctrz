package logging

import (
	"ctrz/utils"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func ProcessLogs(name string, proc *exec.Cmd, detach bool) (cleanup func() error, err error) {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return nil, fmt.Errorf("retrieving ctrz state directory: %w", err)
	}
	path := filepath.Join(dir, "containers", name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}
	logfile, err := os.OpenFile(
		filepath.Join(path, name+".log"),
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o640,
	)
	if err != nil {
		return nil, fmt.Errorf("creating log file: %w", err)
	}

	if detach {
		proc.Stdin = nil
		proc.Stdout = logfile
		proc.Stderr = logfile
		if proc.SysProcAttr == nil {
			proc.SysProcAttr = &syscall.SysProcAttr{}
		}
		proc.SysProcAttr.Setsid = true
	} else {
		proc.Stdout = io.MultiWriter(os.Stdout, logfile)
		proc.Stderr = io.MultiWriter(os.Stderr, logfile)
		proc.Stdin = os.Stdin

	}
	return logfile.Close, nil
}
