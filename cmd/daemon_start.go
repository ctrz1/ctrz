//go:build linux

package cmd

import (
	"fmt"
	"log"

	"ctrz/proc"

	"github.com/spf13/cobra"
)

var startDaemonCmd = &cobra.Command{
	Use:   "start",
	Short: "Start ctrz daemon process",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := proc.StartBackground("daemon")
		if err != nil {
			log.Fatalf("Error starting daemon process: %v\n", err)
		}
		fmt.Printf("Started daemon process: %d\n", pid)
	},
}

func init() {
	daemonCmd.AddCommand(startDaemonCmd)
}
