package cmd

import (
	"ctrz/utils"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon start|stop",
	Short: "control ctrz daemon process",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		utils.EnsureRoot()
		if err := cmd.Usage(); err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
