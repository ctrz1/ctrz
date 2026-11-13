package utils

import (
	"os"
	"path/filepath"
)

var CtrzRootDir = filepath.Join("/var", "lib", "ctrz")

func CtrzStateDir() (string, error) {
	if os.Geteuid() == 0 {
		return CtrzRootDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "ctrz"), nil
}
