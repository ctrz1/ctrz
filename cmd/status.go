package cmd

import (
	"fmt"
	"log"
	"time"
	"bytes"
    "math"
	"strings"

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

		fmt.Printf("PID: %d\n", pid)
		fmt.Printf("Cgroup: %s\n\n", path)
		fmt.Print("\033[s") // Save cursor position

		var lastLines int

		for {
			var buf bytes.Buffer
		
			if ctrls["cpu"] {
				cpu, err := cgroup.ReadCPUStat(path)
				if err != nil {
					fmt.Fprintf(&buf, "CPU: error reading stats: %v\n\n", err)
				} else {
					buf.WriteString("CPU:\n")
					fmt.Fprintf(&buf, "  usage_usec:      %d\n", cpu.UsageUsec)
					fmt.Fprintf(&buf, "  nr_throttled:    %d\n", cpu.NrThrottled)
					fmt.Fprintf(&buf, "  throttled_usec: %d\n\n", cpu.ThrottledUsec)
				}
			}
		
			if ctrls["memory"] {
				mem, err := cgroup.ReadMemStat(path)
				if err != nil {
					fmt.Fprintf(&buf, "Memory: error reading stats: %v\n", err)
				} else {
					buf.WriteString("Memory:\n")
					fmt.Fprintf(&buf, "  current: %d KB\n", mem.Current/1024)
					if mem.Max > math.MaxInt64-1024 || mem.Max == 0 {
						buf.WriteString("  max:     unlimited\n")
					} else {
						fmt.Fprintf(&buf, "  max:     %d KB\n", mem.Max/1024)
					}
				}
			}
		
			output := buf.String()
			lines := strings.Count(output, "\n")
		
			moveCursorUp(lastLines)
			fmt.Print("\033[J")
			fmt.Print(output)
		
			lastLines = lines
			time.Sleep(1 * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Int("pid", 0, "PID of wrapped process")
	statusCmd.MarkFlagRequired("pid")
}

func moveCursorUp(n int) {
	if n > 0 {
		fmt.Printf("\033[%dA", n)
	}
}
