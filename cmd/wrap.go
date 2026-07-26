//go:build linux
// +build linux

package cmd

import (
	"fmt"
	"log"

	"ctrz/cgroup"
	"ctrz/proc"
	"github.com/spf13/cobra"
)

var wrapCmd = &cobra.Command{
	Use:   "wrap",
	Short: "'wrap' a process in a new cgroup",
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := cmd.Flags().GetInt("pid")
		if err != nil {
			log.Fatalf("Unable to retrieve pid: %v", pid)
		}

		if err := cgroup.EnsureCtrzRoot(); err != nil {
			log.Fatalf("cgroup init failed: %v", err)
		}

		info, err := proc.Inspect(pid)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(info.String())
		period, err := cmd.Flags().GetInt("period")
		if err != nil {
			log.Fatal(err)
		}
		runtime, err := cmd.Flags().GetInt("runtime")
		if err != nil {
			log.Fatal(err)
		}
		var maxCpu string
		if period > 0 && runtime > 0 {
			maxCpu = fmt.Sprintf("%d %d", runtime, period)
		} else {
			cpu, err := cmd.Flags().GetInt("cpu")
			if err != nil {
				log.Fatal(err)
			}
			if cpu <= 0 {
				log.Fatal("--cpu flag must have a value bigger than 0")
			}
			quota := cpu * 1000
			maxCpu = fmt.Sprintf("%d 100000", quota)
		}
		if err := cgroup.CreateAndAttach(pid, maxCpu); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(wrapCmd)
	wrapCmd.Flags().Int("pid", 0, "Process ID of process that you want to 'wrap' in a container")
	wrapCmd.MarkFlagRequired("pid")
	wrapCmd.Flags().Int("cpu", 100, "Limits the CPU usage (in %) of the wrapped process")
	wrapCmd.Flags().Int("runtime", 0, "Configures the runtime of a process within a cgroup")
	wrapCmd.Flags().Int("period", 0, "Determines the time window in which to apply the runtime")
	wrapCmd.MarkFlagsRequiredTogether("runtime", "period")
	wrapCmd.MarkFlagsMutuallyExclusive("cpu", "runtime")
}
