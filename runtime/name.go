package runtime

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ctrz/cgroup"
)

func AttachNameToPID(pid int, name string, args []string, containerIP string, containerPort, hostPort []int) error {
	path, err := CtrzStateDir()
	if err != nil {
		return fmt.Errorf("Error attaching name to PID: %v", err)
	}
	err = os.MkdirAll(filepath.Join(path, "containers", name), 0o755)
	if err != nil {
		return fmt.Errorf("Error attaching name to PID: %v", err)
	}
	var command string
	for _, v := range args {
		command += fmt.Sprintf("%s ", v)
	}
	command = strings.Trim(command, " ")
	cgroup, err := cgroup.PathForPID(pid)
	if err != nil {
		return err
	}
	meta := ContainerMeta{
		PID:           pid,
		Name:          name,
		Command:       command,
		Cgroup:        cgroup,
		StartTime:     time.Now().Unix(),
		ContainerIP:   containerIP,
		ContainerPort: containerPort,
		HostPort:      hostPort,
	}
	metaJson, err := json.MarshalIndent(meta, "", "  ")
	fmt.Printf("Container name: %s\n", name)
	if err != nil {
		fmt.Println("Error")
		return err
	}
	err = os.WriteFile(filepath.Join(path, "containers", name, fmt.Sprintf("%s.json", name)), metaJson, 0o644)
	if err != nil {
		return fmt.Errorf("Error attaching name to PID: %v", err)
	}
	return nil
}

func CtrzStateDir() (string, error) {
	if os.Geteuid() == 0 {
		return filepath.Join("/var", "lib", "ctrz"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "ctrz"), nil
}

func GetPIDFromName(name string) (int, error) {
	containerData, err := GetContainerDataFromName(name)
	if err != nil {
		return 0, fmt.Errorf("Error getting PID: %v", err)
	}
	return containerData.PID, nil
}

func CheckContName(name string) bool {
	path, err := CtrzStateDir()
	if err != nil {
		log.Fatalf("Error attaching name to PID: %v", err)
	}
	if _, err := os.Stat(filepath.Join(path, "containers", name, fmt.Sprintf("%s.json", name))); err == nil {
		return false
	}
	return true
}

func GenerateRandomContName() string {
	randId := rand.Int()
	return fmt.Sprintf("ctrz-%d", randId)
}

func RetrieveAllContainers() ([]string, error) {
	stateDir, err := CtrzStateDir()
	if err != nil {
		return nil, err
	}
	dirs, err := os.ReadDir(filepath.Join(stateDir, "containers"))
	if err != nil {
		log.Fatal(err)
	}
	var containers []string
	for _, dir := range dirs {
		if dir.Type().IsDir() {
			containers = append(containers, dir.Name())
		}
	}
	return containers, nil
}
