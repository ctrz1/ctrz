//go:build linux
// +build linux

package cmd

import (
	"fmt"
	"log"

	"ctrz/config"
	"ctrz/runtime"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a new process in an isolated container",
	Run: func(cmd *cobra.Command, args []string) {
		conf, confPath := getFlags(cmd)
		conf.Command = &args
		container, err := config.Get(&conf, confPath)
		if err != nil {
			log.Fatal(err)
		}
		r, err := runtime.New()
		if err != nil {
			log.Fatal(err)
		}
		if err := r.Run(&container); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolP("detach", "d", false, "detach process from terminal")
	runCmd.Flags().Int("cpu", 100, "Limits the CPU usage (in %) of the wrapped process")
	runCmd.Flags().Int("runtime", 0, "Configures the runtime of a process within a cgroup")
	runCmd.Flags().Int("period", 0, "Determines the time window in which to apply the runtime")
	runCmd.Flags().String("name", "", "name of new container")
	runCmd.Flags().StringArrayP("port", "p", []string{}, "Map host port to container port with '<host-port>:<container-port>'")
	runCmd.Flags().Bool("rm", false, "Container automatically cleans up after finishing")
	runCmd.Flags().String("config", "ctrz.yaml", "Path to container config")
	runCmd.MarkFlagsRequiredTogether("runtime", "period")
	runCmd.MarkFlagsMutuallyExclusive("cpu", "runtime")
}

func getFlags(cmd *cobra.Command) (config.CLIOptions, string) {
	var c config.CLIOptions
	detach, err := cmd.Flags().GetBool("detach")
	if err != nil {
		log.Fatal("error retrieving --detach (-d):", err)
	}
	if cmd.Flags().Changed("detach") {
		c.Detach = &detach
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		log.Fatal("error retrieving --name:", err)
	}
	if cmd.Flags().Changed("name") {
		c.Name = &name
	}

	remove, err := cmd.Flags().GetBool("rm")
	if err != nil {
		log.Fatal("error retrieving --rm:", err)
	}
	if cmd.Flags().Changed("remove") {
		c.Remove = &remove
	}

	pm, err := cmd.Flags().GetStringArray("port")
	if err != nil {
		log.Fatal("error retrieving --port (-p)", err)
	}
	if cmd.Flags().Changed("port") {
		c.Ports = &pm
	}

	conf, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatal("error retrieving --config:", err)
	}

	period, err := cmd.Flags().GetInt("period")
	if err != nil {
		log.Fatal(err)
	}
	rt, err := cmd.Flags().GetInt("runtime")
	if err != nil {
		log.Fatal(err)
	}
	var maxCpu string
	if period > 0 && rt > 0 {
		maxCpu = fmt.Sprintf("%d %d", rt, period)
		c.CPU = &maxCpu
	} else {
		cpu, err := cmd.Flags().GetInt("cpu")
		if err != nil {
			log.Fatal(err)
		}
		if cpu <= 0 {
			log.Fatal("--cpu flag must have a value bigger than 0")
		}
		quota := cpu * 1000
		maxCpu = fmt.Sprintf("%d 100000", quota)
		if cmd.Flags().Changed("cpu") {
			c.CPU = &maxCpu
		}
	}

	return c, conf
}
