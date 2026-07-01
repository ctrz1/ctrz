//go:build linux
// +build linux

package cmd

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"ctrz/cgroup"
	"ctrz/network"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show live resource usage of a process",
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := cmd.Flags().GetInt("pid")
		if err != nil {
			log.Fatal(err)
		}
		path, err := cgroup.PathForPID(pid)
		if err != nil {
			log.Fatal(err)
		}
		ctrls, err := cgroup.EnabledControllers(path)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("PID: %d\n", pid)
		fmt.Printf("Cgroup: %s\n\n", path)
		fmt.Print("\033[s") // Save cursor position

		var lastLines int
		var prevRecBytes uint64 = 0
		var prevSentBytes uint64 = 0

		for {
			var buf bytes.Buffer
		
			if ctrls["cpu"] {
				cpu, err := cgroup.ReadCPUStat(path)
				if err != nil {
					fmt.Fprintf(&buf, "CPU: error reading stats: %v\n\n", err)
				} else {
					buf.WriteString("CPU:\n")
					fmt.Fprintf(&buf, "  usage_usec:		%d\n", cpu.UsageUsec)
					fmt.Fprintf(&buf, "  nr_throttled:		%d\n", cpu.NrThrottled)
					fmt.Fprintf(&buf, "  throttled_usec:	%d\n\n", cpu.ThrottledUsec)
				}
			}
		
			if ctrls["memory"] {
				mem, err := cgroup.ReadMemStat(path)
				if err != nil {
					fmt.Fprintf(&buf, "Memory: error reading stats: %v\n", err)
				} else {
					buf.WriteString("Memory:\n")
					fmt.Fprintf(&buf, "  current:			%d KB\n", mem.Current/1024)
					if mem.Max > math.MaxInt64-1024 || mem.Max == 0 {
						buf.WriteString("  max:     unlimited\n")
					} else {
						fmt.Fprintf(&buf, "  max:			%d KB\n", mem.Max/1024)
					}
				}
			}

			sockets, err := network.ResolveSockets(pid)
			if err != nil {
				fmt.Fprintf(&buf, "Network: error reading stats: %v\n\n", err)
			} else {
				currentSent, err := strconv.ParseUint(sockets[0].ReceivedBytes, 10, 64)
				if err != nil {
					
				}
				currentReceived, err := strconv.ParseUint(sockets[0].SentBytes, 10, 64)
				if err != nil {
					
				}
				deltaSent := currentSent - prevSentBytes
				deltaReceived := currentReceived - prevRecBytes
				buf.WriteString("Network:\n")
				fmt.Fprintf(&buf, "  Total Connections:		%d\n", len(sockets))
				fmt.Fprintf(&buf, "  Network traffic (namespace level):\n")
				fmt.Fprintf(&buf, "    Bytes Received:    	%s/second (%s total)\n", convertBytesToFittingUnit(deltaReceived), convertBytesToFittingUnit(currentReceived))
				fmt.Fprintf(&buf, "    Bytes Sent:     	%s/second (%s total)\n\n", convertBytesToFittingUnit(deltaSent), convertBytesToFittingUnit(currentSent))
				connections:
				for i, s := range sockets{
					if s.State == "LISTEN" {
						fmt.Fprintf(&buf, "  %s %s %s\n",s.State, s.Proto, s.RemoteAddr)
					}else {
						fmt.Fprintf(&buf, "  %s %s %s -> %s\n",s.State, s.Proto, s.LocalAddr, s.RemoteAddr)
					}
					fmt.Fprintf(&buf, "  Inode:      		%d\n", s.Inode)
					if i == 5 {
						buf.WriteString("[...]")
						buf.WriteString("\n")
						break connections
					}
					buf.WriteString("\n")
				}
				prevRecBytes = currentReceived
				prevSentBytes = currentSent
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

func convertBytesToFittingUnit(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}