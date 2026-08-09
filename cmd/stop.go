package cmd

import (
	"ctrz/runtime"
	"log"
	"syscall"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop Container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		pid, err := runtime.GetPIDFromName(containerName)
		if err != nil {
			log.Fatalf("Error getting PID: %v\n", err)
		}
		if err := syscall.Kill(pid, syscall.SIGSTOP); err != nil {
			log.Fatalf("Error stopping container: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
