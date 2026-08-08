//go:build linux
// +build linux

package cmd

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ctrz/cgroup"
	"ctrz/network"
	"ctrz/runtime"
	"ctrz/utils"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <container name>",
	Short: "Show live resource usage of a process",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		output, err := cmd.Flags().GetString("out")
		if err != nil {
			log.Fatal(err)
		}
		noHeader, err := cmd.Flags().GetBool("no-header")
		if err != nil {
			log.Fatal(err)
		}
		if output != "" {
			path, err := utils.CtrzStateDir()
			if err != nil {
				log.Fatal(err)
			}
			statf := filepath.Join(path, "containers", containerName, "stats.csv")
			r, err := os.Open(statf)
			if err != nil {
				log.Fatal(err)
			}
			csvr := csv.NewReader(r)
			data, err := csvr.ReadAll()
			if err != nil {
				log.Fatal(err)
			}
			header := "CPU usage (%),CPU usec,CPU nr thorttled,CPU throttled usec,current memory usage (KB),max memory usage (KB),recieved network traffic (delta),received network traffic (total),sent network traffic (delta),sent network traffic (total)"
			if output == "-" {
				if !noHeader {
					fmt.Println(header)
				}
				w := csv.NewWriter(os.Stdout)
				if err := w.WriteAll(data); err != nil {
					log.Fatal(err)
				}
			} else {
				if _, err := os.Stat(output); err == nil {
					log.Fatal("The chosen output file already exists. If you want to append stats to a file, please use 'ctrz status [name] -o - >> someFile.csv'")
				}
				f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o755)
				if err != nil {
					log.Fatal(err)
				}
				w := csv.NewWriter(f)
				if !noHeader {
					fmt.Fprintf(f, "%s\n", header)
				}
				if err := w.WriteAll(data); err != nil {
					log.Fatal(err)
				}
			}
			return
		}

		containerData, err := runtime.GetContainerDataFromName(containerName)
		if err != nil {
			log.Fatal(err)
		}
		path, err := cgroup.PathForPID(containerData.PID)
		if err != nil {
			log.Fatal(err)
		}
		ctrls, err := cgroup.EnabledControllers(path)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\033[3J\033[H\033[2J") // Clear the screen
		fmt.Printf("\033[3J")              // Clear scrollback buffer

		fmt.Printf("PID: %d\n", containerData.PID)
		fmt.Printf("Cgroup: %s\n\n", path)
		fmt.Print("\033[s") // Save cursor position

		var lastLines int
		var prevRecBytes uint64 = 0
		var prevSentBytes uint64 = 0

		var prevUsec uint64 = 0
		prevTime := time.Now()

		for {
			var buf bytes.Buffer

			if ctrls["cpu"] {
				cpu, err := cgroup.ReadCPUStat(path)
				if err != nil {
					fmt.Fprintf(&buf, "CPU: error reading stats: %v\n\n", err)
				} else {
					now := time.Now()
					buf.WriteString("CPU:\n")
					fmt.Fprintf(&buf, "  unsage:		%.2f%%\n", float64(cpu.UsageUsec-prevUsec)/float64(now.Sub(prevTime).Microseconds())*100)
					fmt.Fprintf(&buf, "  usage_usec:		%d\n", cpu.UsageUsec)
					fmt.Fprintf(&buf, "  nr_throttled:		%d\n", cpu.NrThrottled)
					fmt.Fprintf(&buf, "  throttled_usec:	%d\n\n", cpu.ThrottledUsec)

					prevUsec = cpu.UsageUsec
					prevTime = now
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

			sockets, err := network.ResolveSockets(containerData.PID)
			if err != nil {
				fmt.Fprintf(&buf, "Network: error reading stats: %v\n\n", err)
			} else if len(sockets) > 0 {
				currentSent, err := strconv.ParseUint(sockets[0].ReceivedBytes, 10, 64)
				if err != nil {
					currentSent = 0
				}
				currentReceived, err := strconv.ParseUint(sockets[0].SentBytes, 10, 64)
				if err != nil {
					currentReceived = 0
				}
				deltaSent := currentSent - prevSentBytes
				deltaReceived := currentReceived - prevRecBytes
				buf.WriteString("Network:\n")
				fmt.Fprintf(&buf, "  Total Connections:		%d\n", len(sockets))
				fmt.Fprintf(&buf, "  Network traffic (namespace level):\n")
				fmt.Fprintf(&buf, "    Bytes Received:    	%s/second (%s total)\n", convertBytesToFittingUnit(deltaReceived), convertBytesToFittingUnit(currentReceived))
				fmt.Fprintf(&buf, "    Bytes Sent:     	%s/second (%s total)\n\n", convertBytesToFittingUnit(deltaSent), convertBytesToFittingUnit(currentSent))
			connections:
				for i, s := range sockets {
					if s.State == "LISTEN" {
						fmt.Fprintf(&buf, "  %s %s %s\n", s.State, s.Proto, s.RemoteAddr)
					} else {
						fmt.Fprintf(&buf, "  %s %s %s -> %s\n", s.State, s.Proto, s.LocalAddr, s.RemoteAddr)
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
	statusCmd.Flags().StringP("out", "o", "", "Specify output path to get container stats")
	statusCmd.Flags().Bool("no-header", false, "Omits the csv header line")
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
