package runtime

import (
	"ctrz/spec"
	"ctrz/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type pid struct {
	pid       int
	starttime uint64
}

func GetContainerDataFromName(name string) (spec.Container, error) {
	var containerData spec.Container
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
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return containerData, err
	}
	containerData, err = os.ReadFile(filepath.Join(dir, "containers", name, fmt.Sprintf("%s.json", name)))
	if err != nil {
		return containerData, err
	}
	return containerData, nil
}

func getContainerFromPID(dir string) []pid {
	var pids []pid
	b, err := os.ReadFile(filepath.Join(dir, "pids"))
	if err != nil {
		return pids
	}
	data := string(b)
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		fields := strings.Split(line, " ")
		if len(fields) != 2 {
			continue
		}
		p, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		start, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		pids = append(pids, pid{
			pid:       p,
			starttime: start,
		})
	}
	return pids
}
