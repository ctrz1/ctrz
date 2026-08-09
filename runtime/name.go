package runtime

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"ctrz/spec"
	"ctrz/utils"
)

func Name(name string) (string, error) {
	if name == "" {
		name = GenerateRandomContName()
		for !CheckContName(name) {
			name = GenerateRandomContName()
		}
	} else {
		if !CheckContName(name) {
			return "", fmt.Errorf("Container '%s' already exists. Either choose a different name or remove the existing container", name)
		}
	}
	return name, nil
}

func AttachNameToPID(container spec.Container) error {
	path, err := utils.CtrzStateDir()
	if err != nil {
		return fmt.Errorf("Error attaching name to PID: %v", err)
	}
	err = os.MkdirAll(filepath.Join(path, "containers", container.Spec.Name), 0o755)
	if err != nil {
		return fmt.Errorf("Error attaching name to PID: %v", err)
	}
	containerJson, err := json.MarshalIndent(container, "", "  ")
	fmt.Printf("Container name: %s\n", container.Spec.Name)
	if err != nil {
		fmt.Println("Error")
		return err
	}
	err = os.WriteFile(filepath.Join(path, "containers", container.Spec.Name, fmt.Sprintf("%s.json", container.Spec.Name)), containerJson, 0o644)
	if err != nil {
		return fmt.Errorf("Error attaching name to PID: %v", err)
	}
	return nil
}

func GetPIDFromName(name string) (int, error) {
	containerData, err := GetContainerDataFromName(name)
	if err != nil {
		return 0, fmt.Errorf("Error getting PID: %v", err)
	}
	return containerData.PID, nil
}

func CheckContName(name string) bool {
	path, err := utils.CtrzStateDir()
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
	stateDir, err := utils.CtrzStateDir()
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
