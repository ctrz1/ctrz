package cmd

import (
	"ctrz/utils"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:    "dashboard start|stop",
	Short:  "expose ctrz container metrics via admin dashboard",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		utils.EnsureRoot()
		if err := cmd.Usage(); err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
