package logging

import (
	"os"
	"os/exec"
)

func ProcessLogs(name string, proc *exec.Cmd, detach bool) error {
	//dir, err := CtrzStateDir()
	//if err != nil {
	//	return fmt.Errorf("Error retrieving ctrz state directory: %v", err)
	//}
	//logfile, err := os.Create(filepath.Join(dir, "containers", fmt.Sprintf("%s.log", name)))
	//if err != nil {
	//	return fmt.Errorf("Error creating log file: %v", err)
	//}
	//defer logfile.Close()
	if !detach {
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		proc.Stdin = os.Stdin
	}
	return nil
}
