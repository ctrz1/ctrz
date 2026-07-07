package cmd

import (
	"ctrz/misc"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon [start/stop]",
	Short: "Start/Stop ctrz daemon process",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]
		dir, err := misc.CtrzStateDir()
		path := filepath.Join(dir, "daemon")

		if err != nil {
			log.Fatal(err)
		}
		
		switch action {
		case "start":
			data, err := os.ReadFile(path)
			if err == nil {
				log.Fatalf("Daemon process already running with PID: %s\n", string(data))
			}

			d := exec.Command("/proc/self/exe", "__deamon", action)

			if err := d.Start(); err != nil {
				log.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(strconv.Itoa(d.Process.Pid)), 0755); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Started daemon process: %d\n", d.Process.Pid)
			
		case "stop":
			data, err := os.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}
			pid := string(data)
			if _, err := os.Open(fmt.Sprintf("/proc/%s/stat", pid)); err == nil {
				out, err := exec.Command("kill", pid).CombinedOutput()
				if err != nil {
					log.Fatalf("%v: %s", err, out)
				}
			}
			if err := os.Remove(path); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
