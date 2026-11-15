package cmd

import (
	"ctrz/proc"
	"log"

	"github.com/spf13/cobra"
)

var dashboardStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop ctrz admin dashboard",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := proc.CleanBackground("dashboard"); err != nil {
			log.Fatalf("Error cleaning up admin dashboard: %v", err)
		}
	},
}

func init() {
	dashboardCmd.AddCommand(dashboardStopCmd)
}
