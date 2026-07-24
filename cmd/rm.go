//go:build linux
// +build linux

package cmd

import (
	"ctrz/misc"
	"ctrz/proc"
	"log"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a containern and all of its data (including stats)",

	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("unable to retrieve name: %v\n", err)
		}
		forceKill, err := cmd.Flags().GetBool("force")
		if err != nil {
			log.Fatalf("unable to retrieve force: %v\n", err)
		}
		inactive, err := cmd.Flags().GetBool("inactive")
		if err != nil {
			log.Fatalf("unable to retrieve inactive: %v\n", err)
		}
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Fatalf("unable to retrieve all: %v\n", err)
		}
		if inactive || all {
			containers, err := misc.RetrieveAllContainers()
			if err != nil {
				log.Fatalf("Could not retrieve list of containers: %v\n", err)
			}
			for _, c := range containers {
				containerMeta, err := misc.GetContainerDataFromName(c)
				if err != nil {
					continue
				}
				if !proc.IsProcActive(containerMeta.PID) || all {
					misc.RemoveContainerByName(c, forceKill)
				}
			}
			return
		}
		if err := misc.RemoveContainerByName(name, forceKill); err != nil {
			log.Fatalf("Error cleaning up container: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().String("name", "", "Name of the container to be removed")
	rmCmd.Flags().BoolP("force", "9", false, "Uses KILL instead TERM command. This forces a process to stop immediately and does not allow for cleanup routines")
	rmCmd.Flags().Bool("inactive", false, "Removes all inactive containers")
	rmCmd.Flags().BoolP("all", "a", false, "Remove all containers")
	rmCmd.MarkFlagsMutuallyExclusive("name", "inactive", "all")
}
