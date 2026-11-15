package logging

import (
	"bufio"
	"context"
	"ctrz/utils"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
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

func GetLogs(name string, w io.Writer, tailBytes int64) (*os.File, error) {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return nil, fmt.Errorf("retrieving ctrz state directory: %w", err)
	}

	f, err := os.Open(filepath.Join(dir, "containers", name, name+".log"))
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	if tailBytes > 0 {
		size, err := f.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		if size > tailBytes {
			if err := skipToLineStart(f, size-tailBytes); err != nil {
				return nil, err
			}
		} else if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	_, err = io.Copy(w, f)
	return f, err
}

func skipToLineStart(f *os.File, offset int64) error {
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	br := bufio.NewReader(f)
	line, err := br.ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	_, err = f.Seek(offset+int64(len(line)), io.SeekStart)
	return err
}

func AttachLogs(ctx context.Context, f *os.File, w io.Writer) error {
	buf := make([]byte, 32*1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if errors.Is(err, io.EOF) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
			continue
		}
		if err != nil {
			return err
		}
	}
}
