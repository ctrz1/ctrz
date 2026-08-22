package config

import (
	"ctrz/spec"
	"testing"
)

func TestMerge(t *testing.T) {
	name := "testcontainer"
	cli := CLIOptions{
		Name:    &name,
		Command: &[]string{"hello", "world"},
	}
	conf := spec.ContainerSpec{
		Name:    "config",
		Command: []string{"other", "command"},
		Remove:  true,
	}
	expected := spec.ContainerSpec{
		Name:    name,
		Command: []string{"hello", "world"},
		Remove:  true,
		Detach:  false,
		Ports:   []string{},
		CPU:     "100000 100000",
	}
	actual := merge(&cli, &conf)
	if !equals(actual, expected) {
		t.Fatalf("got %+v \nexpected: %+v", actual, expected)
	}
}

func equals(a, b spec.ContainerSpec) bool {
	if a.CPU != b.CPU || a.Detach != b.Detach || a.Name != b.Name || a.Remove != b.Remove {
		return false
	}
	for i, v := range a.Command {
		if b.Command[i] != v {
			return false
		}
	}
	for i, v := range a.Ports {
		if b.Ports[i] != v {
			return false
		}
	}
	return true
}
