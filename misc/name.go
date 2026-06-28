package misc

import (
	"ctrz/cgroup"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

//TODO:
/**
1. Make names unique --> if a name is chosen again, check if the process is stale or not (either replace process or throw an error)
	- Check if /proc/<pid> exists
	- Compare StartTime from /var/lib/ctrz/containers/xyz.json to /proc/<pid>/stat
2. Implement cleanup ('rm' command)
3. Integrate names into status & wrap command
4. Implement a 'ps' command showing active processes
**/

func AttachNameToPID(pid int, name string, args []string, containerIP string, containerPort []int, hostPort []int) error {
	path, err := ctrzStateDir()
	if err != nil {
		return fmt.Errorf("Error attaching name to PID: %v", err)
	}
	err = os.MkdirAll(filepath.Join(path, "containers", name), 0755)
	if err != nil {
		return fmt.Errorf("Error attaching name to PID: %v", err)
	}
	// TODO: Include process start time here
	var command string
	for _, v := range args {
		command += fmt.Sprintf("%s ", v)
	}
	command = strings.Trim(command, " ")
	cgroup, err := cgroup.PathForPID(pid)
	if err != nil {
		return err
	}
	meta := ContainerMeta {
		PID: pid,
		Name: name,
		Command: command,
		Cgroup: cgroup,
		ContainerIP: containerIP,
		ContainerPort: containerPort,
		HostPort: hostPort,
	}
	metaJson, err := json.Marshal(meta)
	fmt.Printf("%v\n", string(metaJson))
	if err != nil {
		fmt.Println("Error")
		return err
	}
	err = os.WriteFile(filepath.Join(path, "containers", name, fmt.Sprintf("%s.json", name)), metaJson, 0644)
	if err != nil {
		return fmt.Errorf("Error attaching name to PID: %v", err)
	}
	return nil
}

func ctrzStateDir() (string, error) {
	if os.Geteuid() == 0 {
		return filepath.Join("/var", "lib", "ctrz"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "ctrz"), nil
}


func GetPIDFromName(name string) (int, error){
	//TODO
	return 0, nil
}

func CheckContName(name string) bool {
	path, err := ctrzStateDir()
	if err != nil {
		log.Fatalf("Error attaching name to PID: %v", err)
	}
	if _, err := os.Stat(filepath.Join(path, "containers", fmt.Sprintf("%s.json", name))); err == nil {
		return false
	}
	return true
}

func GenerateRandomContName() string {
	randId := rand.Int()
	return fmt.Sprintf("ctrz-%d", randId)
}