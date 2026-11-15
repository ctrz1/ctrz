//go:build linux

package cmd

import (
	"context"
	"ctrz/logging"
	"log"
	"os"

	"github.com/spf13/cobra"
)

const MiB int64 = 1<<20

var logCmd = &cobra.Command{
	Use:   "log [name]",
	Short: "Print the log file of a container",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		attach, err := cmd.Flags().GetBool("attach")
		if err != nil {
			log.Fatalf("Unable to retrive --attach: %v", err)
		}
		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			log.Fatalf("Unable to retrive --limit (-l): %v", err)
		}

		name := args[0]
		f, err := logging.GetLogs(name, os.Stdout, int64(limit)*MiB)
		if err != nil {
			log.Fatalf("Error getting logs: %v", err)
		}

		if attach {
			if err := logging.AttachLogs(context.TODO(), f, os.Stdout); err != nil {
				log.Fatalf("Error attaching logs to terminal: %v", err)
			}
		}
		f.Close()
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().Bool("attach", false, "Stream container logs to terminal")
	logCmd.Flags().IntP("limit", "l", 1, "Size limit before the logfile gets truncated")
}
