package cgroup

import (
	"os"
	"strings"
)

/**
func EnableControllers(controllers string) error {
	path := "/sys/fs/cgroup/cgroup.subtree_control"
	return os.WriteFile(path, []byte(controllers), 0644)
}
**/

func AvailableControllers() (map[string]bool, error) {
	data, err := os.ReadFile("/sys/fs/cgroup/cgroup.controllers")
	if err != nil {
		return nil, err
	}
	ctrls := make(map[string]bool)
	for _, c := range strings.Fields(string(data)) {
		ctrls[c] = true
	}
	return ctrls, nil
}

