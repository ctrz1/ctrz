package misc

type ContainerMeta struct {
	PID       int
	Name      string
	StartTime int64
	Command   string
	Cgroup    string
}