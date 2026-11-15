//go:build linux

package cmd

import (
	"fmt"
	"log"

	"ctrz/proc"

	"github.com/spf13/cobra"
)

var startDashboardCmd = &cobra.Command{
	Use:   "start",
	Short: "Start ctrz admin dashboard",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := proc.StartBackground("dashboard"); err != nil {
			log.Fatalf("Error starting admin dashboard: %v\n", err)
		}
		fmt.Printf("Started dashboard\n")
	},
}

func init() {
	dashboardCmd.AddCommand(startDashboardCmd)
}
