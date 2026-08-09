package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"ctrz/cgroup"
	"ctrz/proc"
	"ctrz/utils"

	"github.com/spf13/cobra"
)

var startDaemonCmd = &cobra.Command{
	Use:   "start",
	Short: "Start ctrz daemon process",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := utils.CtrzStateDir()
		path := filepath.Join(dir, "daemon")

		if err != nil {
			log.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err == nil {
			s := strings.Split(string(data), " ")
			pid, err := strconv.Atoi(s[0])
			if err != nil {
				log.Fatal(err)
			}
			starttime, err := strconv.ParseUint(s[1], 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			if proc.IsProcActive(pid, starttime) {
				log.Fatalf("Daemon process already running with PID: %d\n", pid)
			}
		}

		d := exec.Command("/proc/self/exe", "__deamon")

		if err := d.Start(); err != nil {
			log.Fatal(err)
		}
		cgroupManager := cgroup.New()
		if err := cgroupManager.CreateAndAttach(d.Process.Pid, "100000 100000"); err != nil {
			log.Fatal(err)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatal(err)
		}
		stat, err := proc.ProcessStats(d.Process.Pid)
		if err != nil {
			log.Fatal(err)
		}
		daemonInfo := fmt.Sprintf("%d %d", d.Process.Pid, stat.Starttime)
		if err := os.WriteFile(path, []byte(daemonInfo), 0o755); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Started daemon process: %d\n", d.Process.Pid)
	},
}

func init() {
	daemonCmd.AddCommand(startDaemonCmd)
}
