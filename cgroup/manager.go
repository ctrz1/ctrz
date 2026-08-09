package cgroup

import (
	"os"
	"path/filepath"
	"strings"
)

const CtrzRoot = "/sys/fs/cgroup/ctrz"

type Manager struct {
	Root string
}

func New() Manager {
	return Manager{
		Root: CtrzRoot,
	}
}

func EnsureCtrzRoot() error {
	if _, err := os.Stat(CtrzRoot); os.IsNotExist(err) {
		if err := os.Mkdir(CtrzRoot, 0o755); err != nil {
			return err
		}
	}

	ctrls, err := AvailableControllers()
	if err != nil {
		return err
	}

	var enable []string
	if ctrls["cpu"] {
		enable = append(enable, "+cpu")
	}
	if ctrls["memory"] {
		enable = append(enable, "+memory")
	}

	if len(enable) == 0 {
		return nil
	}

	return os.WriteFile(
		filepath.Join(CtrzRoot, "cgroup.subtree_control"),
		[]byte(strings.Join(enable, " ")),
		0o644,
	)
}
