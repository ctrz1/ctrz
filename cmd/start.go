package cmd

import (
	"ctrz/runtime"
	"log"
	"syscall"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start Container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		//GetPIDFromName does not work
		pid, err := runtime.GetPIDFromName(containerName)
		if err != nil {
			log.Fatalf("Error getting PID: %v\n", err)
		}
		if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
			log.Fatalf("Error starting container: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
