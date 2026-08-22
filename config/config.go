package config

import (
	"ctrz/spec"
	"fmt"
	"log/slog"
)

type CLIOptions struct {
	Name    *string   `json:"name"`
	CPU     *string   `json:"cpu"`
	Command *[]string `json:"command"`
	Remove  *bool     `json:"remove"`
	Detach  *bool     `json:"detached"`
	Ports   *[]string `json:"ports"`
}

func Get(cont *CLIOptions, confPath string) (spec.ContainerSpec, error) {

	c, err := load(confPath)
	if err != nil && confPath != defaultConfPath{
		slog.Warn(fmt.Sprintf("Unable to load config at '%s': %v", confPath, err))
	}

	return merge(cont, c), nil
}
