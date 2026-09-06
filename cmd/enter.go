package cmd

import (
	"ctrz/runtime"
	"ctrz/utils"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

var enterCmd = &cobra.Command{
	Use:   "enter <container name> [OPTIONS]",
	Short: "Enter into an interactive shell within a container",
	Run: func(cmd *cobra.Command, args []string) {
		utils.EnsureRoot()
		command, err := cmd.Flags().GetString("command")
		if err != nil {
			log.Fatal(err)
		}
		containerName := args[0]
		containerData, err := runtime.GetContainerDataFromName(containerName)
		if err != nil {
			log.Fatalf("Error retrieving container data: %v\n", err)
		}
		if command == "" {
			command = "sh"
		}
		commandArr := strings.Split(command, " ")
		path, err := exec.LookPath("nsenter")
		if err != nil {
			log.Fatal("nsenter not found in PATH")
		}

		if err := unix.Exec(
			path,
			append([]string{"nsenter", "-a", "-t", strconv.Itoa(containerData.PID)}, commandArr...),
			os.Environ(),
		); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(enterCmd)
	enterCmd.Flags().StringP("command", "c", "", "Pass a command to the container without entering. This flag makes the command non-interactive")
}
