package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"ctrz/cgroup"
	"ctrz/proc"
	"ctrz/runtime"

	"github.com/spf13/cobra"
)

var startDaemonCmd = &cobra.Command{
	Use:   "start",
	Short: "Start ctrz daemon process",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := runtime.CtrzStateDir()
		path := filepath.Join(dir, "daemon")

		if err != nil {
			log.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err == nil {
			pid, err := strconv.Atoi(string(data))
			if err != nil {
				log.Fatal(err)
			}
			if proc.IsProcActive(pid) {
				log.Fatalf("Daemon process already running with PID: %d\n", pid)
			}
		}

		d := exec.Command("/proc/self/exe", "__deamon")

		if err := d.Start(); err != nil {
			log.Fatal(err)
		}
		if err := cgroup.CreateAndAttach(d.Process.Pid, "100000 100000"); err != nil {
			log.Fatal(err)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(strconv.Itoa(d.Process.Pid)), 0o755); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Started daemon process: %d\n", d.Process.Pid)
	},
}

func init() {
	daemonCmd.AddCommand(startDaemonCmd)
}
