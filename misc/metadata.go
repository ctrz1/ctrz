package misc

type ContainerMeta struct {
	PID       int    `json:"pid"`
	Name      string `json:"name"`
	StartTime int64  `json:"startTime,omitempty"`
	Command   string `json:"command"`
	Cgroup    string `json:"cgroup"`
}
