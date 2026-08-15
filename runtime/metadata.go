package runtime

import (
	"ctrz/spec"
	"ctrz/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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
