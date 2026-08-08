package spec

type PortMapping struct {
	HostPort      int `json:"hostPort"`
	ContainerPort int `json:"containerPort"`
}

type Network struct {
	Ports []PortMapping `json:"ports"`
	IP    string        `json:"IP"`
}
