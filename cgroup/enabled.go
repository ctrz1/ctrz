package cgroup

import (
	"os"
	"path/filepath"
	"strings"
)

func EnabledControllers(cgroupPath string) (map[string]bool, error) {
	data, err := os.ReadFile(filepath.Join(cgroupPath, "/cgroup.controllers"))
	if err != nil {
		return nil, err
	}

	ctrls := make(map[string]bool)
	for _, c := range strings.Fields(string(data)) {
		ctrls[c] = true
	}
	return ctrls, nil
}
