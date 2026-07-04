package cmd

import (
	"ctrz/misc"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

var enterCmd = &cobra.Command{
	Use:   "enter <container name>",
	Short: "Enter into an interactive shell within a container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		containerData, err := misc.GetContainerDataFromName(containerName)
		if err != nil {
			log.Fatalf("Error retrieving container data: %v\n", err)
		}
		shell := "sh"
		path, err := exec.LookPath("nsenter")
		if err != nil {
		    log.Fatal("nsenter not found in PATH")
		}

		if err := unix.Exec(
		    path,
		    []string{"nsenter", "-a", "-t", strconv.Itoa(containerData.PID), shell},
		    os.Environ(),
		); err != nil {
		    log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(enterCmd)
}
