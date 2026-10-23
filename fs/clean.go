package fs

import (
	"ctrz/utils"
	"os"
)

func Clean() error {
	dir, err := utils.CtrzStateDir()
	if err != nil {
		return err
	}

	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return nil
}