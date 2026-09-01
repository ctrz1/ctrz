package config

import (
	"ctrz/runtime"
	"ctrz/spec"
	"fmt"
	"log"
)

// This function merges user input from the cli with container config
// from e.g. ctrz.yaml
// User input from the cli superseeds config files
func merge(cli *CLIOptions, conf *spec.ContainerSpec) spec.ContainerSpec {
	var ret spec.ContainerSpec

	// Name
	if cli != nil && cli.Name != nil {
		ret.Name = *cli.Name
	} else if conf != nil && conf.Name != "" {
		ret.Name = conf.Name
	}

	name, err := runtime.Name(ret.Name)
	if err != nil {
		log.Fatal(err)
	}
	ret.Name = name

	// CPU
	if cli != nil && cli.CPU != nil {
		ret.CPU = *cli.CPU
	} else if conf != nil && conf.CPU != "" {
		ret.CPU = fmt.Sprintf("%s000 100000", conf.CPU)
	} else {
		ret.CPU = fmt.Sprintf("%d000 100000", defaultCpu)
	}

	// Command
	if cli != nil && len(*cli.Command) != 0 {
		ret.Command = *cli.Command
	} else if conf != nil && len(conf.Command) != 0 {
		ret.Command = conf.Command
	} else {
		log.Fatal("At least one command must be provided")
	}

	// Remove
	if cli != nil && cli.Remove != nil {
		ret.Remove = *cli.Remove
	} else if conf != nil {
		ret.Remove = conf.Remove
	}

	// Detach
	if cli != nil && cli.Detach != nil {
		ret.Detach = *cli.Detach
	} else if conf != nil {
		ret.Detach = conf.Detach
	}

	// Ports
	if cli != nil && cli.Ports != nil {
		ret.Ports = *cli.Ports
	} else if conf != nil && conf.Ports != nil {
		ret.Ports = conf.Ports
	}

	return ret
}
