package spec

// Actual container
type Container struct {
	PID         int           `json:"pid"`
	Spec        ContainerSpec `json:"containerSpec"`
	StartTime   uint64        `json:"startTime"`
	Started     int64         `json:"started"`
	Cgroup      string        `json:"cgroup"`
	NetworkSpec Network       `json:"network"`
	ProcStats   ProcStats     `json:"process"`
}

// Specification based on user input
type ContainerSpec struct {
	Name    string   `json:"name"`
	CPU     string   `json:"cpu"`
	Command []string `json:"command"`
	Remove  bool     `json:"remove"`
	Detach  bool     `json:"detached"`
	Ports   []string `json:"ports"`
}

type Removal struct {
	Name     string
	Force    bool
	All      bool
	Inactive bool
}
