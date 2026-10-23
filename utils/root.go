package utils

import (
	"fmt"
	"os"
)

func EnsureRoot() {
	if os.Geteuid() != 0 {
		fmt.Println("ctrz currently does not support rootles containers")
		fmt.Println("please execute ctrz with root privileges")
		os.Exit(1)
	}
}
