//go:build linux
// +build linux

package proc

import (
	"fmt"
	"os"
	"path/filepath"
)

type Info struct {
	PID        int
	Namespaces map[string]string
	Cgroup     string
}

func Inspect(pid int) (*Info, error) {
	info := &Info{
		PID:        pid,
		Namespaces: make(map[string]string),
	}

	// Read namespaces
	nsDir := fmt.Sprintf("/proc/%d/ns", pid)
	entries, err := os.ReadDir(nsDir)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		target, err := os.Readlink(filepath.Join(nsDir, e.Name()))
		if err != nil {
			continue
		}
		info.Namespaces[e.Name()] = target
	}

	// Read cgroup
	cg, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err == nil {
		info.Cgroup = string(cg)
	}

	return info, nil
}

func (i *Info) String() string {
	s := fmt.Sprintf("PID: %d\n\nNamespaces:\n", i.PID)
	for k, v := range i.Namespaces {
		s += fmt.Sprintf("  %-8s -> %s\n", k, v)
	}
	s += "\nCgroup:\n" + i.Cgroup
	return s
}
