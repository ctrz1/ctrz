//go:build linux
// +build linux

package misc

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
