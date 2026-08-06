package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"ctrz/proc"
	"ctrz/runtime"

	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Print list of (running) containers",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		containers, err := runtime.RetrieveAllContainers()
		if err != nil {
			log.Fatalf("Error retrieving containers: %v\n", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		defer func() {
			if err := w.Flush(); err != nil {
				log.Fatal(err)
			}
		}()

		if _, err := fmt.Fprintln(w); err != nil {
			log.Fatal(err)
		}
		// fmt.Fprintln(w, "-------------------------------------------------------------------------------")
		if _, err := fmt.Fprintln(w, "Name\tPID\tIP\tCreated\tCommand\tStatus"); err != nil {
			log.Fatal(err)
		}
		// fmt.Fprintln(w, "-------------------------------------------------------------------------------")

		for _, c := range containers {
			containerData, err := runtime.GetContainerDataFromName(c)
			if err != nil {
				fmt.Printf("Error getting container info: %v\n", err)
				continue
			}
			var status string
			cStats, err := proc.ProcessStats(containerData.PID)
			if proc.IsProcActive(containerData.PID, containerData.StartTime) {
				status = "running"
			} else if err == nil {
				status = fmt.Sprintf("stopped (%d)", cStats.Exit_code)
			} else {
				status = "dangling"
			}
			created := time.Unix(containerData.Started, 0).Format("02/01/2006 15:04:05")
			if _, err := fmt.Fprintf(w,
				"%s\t%d\t%s\t%s\t%s\t%s\n",
				containerData.Name,
				containerData.PID,
				containerData.ContainerIP,
				created,
				containerData.Command,
				status,
			); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}
