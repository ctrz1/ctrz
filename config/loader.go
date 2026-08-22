package config

import (
	"ctrz/spec"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Container map[string]spec.ContainerSpec `yaml:"container"`
}

func load(path string) (*spec.ContainerSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading config file: %v\n", err)
	}
	var config Config

	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	var cont spec.ContainerSpec

	for name, container := range config.Container {
		container.Name = name
		config.Container[name] = container
		cont = container
	}

	return &cont, nil
}
