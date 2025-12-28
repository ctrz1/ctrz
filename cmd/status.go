package cmd

import (
	"fmt"
	"log"
	"time"

	"ctrz/cgroup"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show live resource usage of a wrapped process",
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := cmd.Flags().GetInt("pid")
		if err != nil {
			log.Fatal(err)
		}
		path := cgroup.PathForPID(pid)
		ctrls, err := cgroup.EnabledControllers(path)
		if err != nil {
			log.Fatal(err)
		}

		for {
			fmt.Print("\033[H\033[2J") // clear screen

			if ctrls["cpu"] {
				cpu, err := cgroup.ReadCPUStat(path)
				if err != nil {
					fmt.Printf("error reading cpu stats: %v\n", err)
				}

				fmt.Printf("PID: %d\n", pid)
				fmt.Printf("Cgroup: %s\n\n", path)

				fmt.Println("CPU:")
				fmt.Printf("  usage_usec:      %d\n", cpu.UsageUsec)
				fmt.Printf("  nr_throttled:    %d\n", cpu.NrThrottled)
				fmt.Printf("  throttled_usec: %d\n\n", cpu.ThrottledUsec)
			}
			if ctrls["memory"] {
				mem, err := cgroup.ReadMemStat(path)
				if err != nil {
					fmt.Printf("error reading memory stats: %v\n", err)
				}
				fmt.Println("Memory:")
				fmt.Printf("  current: %d KB\n", mem.Current/1024)
				if mem.Max == 0 {
					fmt.Println("  max:     unlimited")
				} else {
					fmt.Printf("  max:     %d KB\n", mem.Max/1024)
				}
			}

			time.Sleep(1 * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Int("pid", 0, "PID of wrapped process")
	statusCmd.MarkFlagRequired("pid")
}
