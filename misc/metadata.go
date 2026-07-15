package misc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ContainerMeta struct {
	PID           int    `json:"pid"`
	Name          string `json:"name"`
	StartTime     int64  `json:"startTime,omitempty"`
	Command       string `json:"command"`
	Cgroup        string `json:"cgroup"`
	ContainerIP   string `json:"ContainerIP"`
	ContainerPort []int  `json:"containerPort,omitempty"`
	HostPort      []int  `json:"hostPort,omitempty"`
}

func GetContainerDataFromName(name string) (ContainerMeta, error) {
	var containerData ContainerMeta
	data, err := GetRawContainerDataFromName(name)
	if err != nil {
		return containerData, err
	}
	if err := json.Unmarshal(data, &containerData); err != nil {
		return containerData, err
	}
	return containerData, nil
}

func GetRawContainerDataFromName(name string) ([]byte, error) {
	var containerData []byte
	dir, err := CtrzStateDir()
	if err != nil {
		return containerData, err
	}
	containerData, err = os.ReadFile(filepath.Join(dir, "containers", name, fmt.Sprintf("%s.json", name)))
	if err != nil {
		return containerData, err
	}
	return containerData, nil
}
