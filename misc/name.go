package misc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func AttachNameToPID(pid int, name string) error {
	path, err := ctrzStateDir()
	if err != nil {
		return err
	}
	os.Mkdir(filepath.Join(path, "containers"), 0755)
	// TODO: Expand on metadata here
	meta := ContainerMeta {
		PID: pid,
		Name: name,
	}
	metaJson, err := json.Marshal(meta)
	fmt.Printf("%v\n", string(metaJson))
	if err != nil {
		fmt.Println("Error")
		return err
	}
	err = os.WriteFile(filepath.Join(path, "containers", fmt.Sprintf("%s.json", name)), metaJson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func ctrzStateDir() (string, error) {
	if os.Geteuid() == 0 {
		return filepath.Join("var", "lib", "ctrz"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "ctrz"), nil
}


func GetPIDFromName(name string) (int, error){
	return 0, nil
}