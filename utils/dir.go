package utils

import (
	"os"
	"path/filepath"
)

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
