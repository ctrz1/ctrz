//go:build linux
// +build linux

package cmd

import (
	"ctrz/network"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use: "run",
	Short: "Run a new process in an isolated container",
	Run: func(cmd *cobra.Command, args []string) {
		netns, err := cmd.Flags().GetBool("netns")
		if err != nil {
			log.Fatal("error retrieving --netns. Setting value to default (false): ", err)
			netns = false
		}
		detach, err := cmd.Flags().GetBool("detach")
		if err != nil {
			log.Fatal("error retrieving --detach (-d). Setting value to default (false): ", err)
			detach = false
		}

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
		if netns {
			fmt.Println("Let's create a new network namespace")
			if len(args) < 1 {
				log.Fatal("At least one command must be provided")
				os.Exit(1)
			}
			err = network.CreateNetNs(args, detach, maxCpu)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("netns", false, "Create a new network namespace")
	runCmd.Flags().BoolP("detach", "d", false, "detach process from terminal")
	runCmd.Flags().Int("cpu", 100, "Limits the CPU usage (in %) of the wrapped process")
	runCmd.Flags().Int("runtime", 0, "Configures the runtime of a process within a cgroup")
	runCmd.Flags().Int("period", 0, "Determines the time window in which to apply the runtime")
	runCmd.MarkFlagsRequiredTogether("runtime", "period")
	runCmd.MarkFlagsMutuallyExclusive("cpu", "runtime")
}