package cmd

import (
	"ctrz/misc"
	"ctrz/proc"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop ctrz daemon process",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := misc.CtrzStateDir()
		path := filepath.Join(dir, "daemon")

		if err != nil {
			log.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		pidS := string(data)
		pid, err := strconv.Atoi(string(data))
		if proc.IsProcActive(pid) {
			out, err := exec.Command("kill", pidS).CombinedOutput()
			if err != nil {
				log.Fatalf("%v: %s", err, out)
			}
		}
		if err := os.Remove(path); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	daemonCmd.AddCommand(daemonStopCmd)
}
