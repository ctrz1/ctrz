package cmd

import (
	"log"

	"ctrz/proc"

	"github.com/spf13/cobra"
)

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop ctrz daemon process",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := proc.CleanBackground("daemon"); err != nil {
			log.Fatalf("Error cleaning up daemon process: %v", err)
		}
	},
}

func init() {
	daemonCmd.AddCommand(daemonStopCmd)
}
