package cgroup

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CPUStat struct {
	UsageUsec     uint64
	NrThrottled   uint64
	ThrottledUsec uint64
}

type MemStat struct {
	Current uint64
	Max     uint64
}

func ReadCPUStat(path string) (*CPUStat, error) {
	data, err := os.ReadFile(filepath.Join(path, "cpu.stat"))
	if err != nil {
		return nil, err
	}

	stat := &CPUStat{}
	lines := strings.Split(string(data), "\n")

	for _, l := range lines {
		parts := strings.Fields(l)
		if len(parts) != 2 {
			continue
		}
		val, _ := strconv.ParseUint(parts[1], 10, 64)
		switch parts[0] {
		case "usage_usec":
			stat.UsageUsec = val
		case "nr_throttled":
			stat.NrThrottled = val
		case "throttled_usec":
			stat.ThrottledUsec = val
		}
	}
	return stat, nil
}

func ReadMemStat(path string) (*MemStat, error) {
	cur, err := os.ReadFile(filepath.Join(path, "memory.current"))
	if err != nil {
		return nil, err
	}
	max, err := os.ReadFile(filepath.Join(path, "memory.max"))
	if err != nil {
		return nil, err
	}

	curVal, err := strconv.ParseUint(strings.TrimSpace(string(cur)), 10, 64)
	if err != nil {
		return nil, err
	}

	var maxVal uint64
	if strings.TrimSpace(string(max)) == "max" {
		maxVal = 0
	} else {
		maxVal, err = strconv.ParseUint(strings.TrimSpace(string(max)), 10, 64)
		if err != nil {
			return nil, err
		}
	}

	return &MemStat{
		Current: curVal,
		Max:     maxVal,
	}, nil
}
