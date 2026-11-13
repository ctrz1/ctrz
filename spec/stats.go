package spec

import "time"

type Stats struct {
	Name      string        `json:"name"`
	Timestamp time.Time     `json:"timestamp"`
	PID       int           `json:"pid"`
	Cgroup    string        `json:"cgroup"`
	CPU       *CPUStats     `json:"cpu,omitempty"`
	Memory    *MemoryStats  `json:"memory,omitempty"`
	Network   *NetworkStats `json:"network,omitempty"`
	Errors    []string      `json:"errors,omitempty"`
	Container Container     `json:"container_spec"`
}

type CPUStats struct {
	UsagePercent  *float64 `json:"usage_percent"`
	UsageUsec     uint64   `json:"usage_usec"`
	NrThrottled   uint64   `json:"nr_throttled"`
	ThrottledUsec uint64   `json:"throttled_usec"`
}

type MemoryStats struct {
	Current string `json:"current"`
	Max     string `json:"max"`
}

type NetworkStats struct {
	TotalConnections  int          `json:"total_connections"`
	ReceivedPerSecond string       `json:"received_per_second"`
	ReceivedTotal     string       `json:"received_total"`
	SentPerSecond     string       `json:"sent_per_second"`
	SentTotal         string       `json:"sent_total"`
	Connections       []Connection `json:"connections"`
}

type Connection struct {
	State      string `json:"state"`
	Proto      string `json:"proto"`
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	Inode      any    `json:"inode"`
}
