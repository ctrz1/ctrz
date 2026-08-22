package config

import (
	"ctrz/spec"
	"testing"
)

func TestLoader(t *testing.T) {
	expected := spec.ContainerSpec{
		Name:    "demo",
		Ports:   []string{"8443:8443"},
		CPU:     "80",
		Detach:  true,
		Command: []string{"demo-webserver"},
	}

	conf, err := load("testdata/example.yaml")
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
	if !equals(expected, *conf) {
		t.Fatalf("got: %+v; expected: %+v\n", *conf, expected)
	}
}

func TestLoadInvalidFile(t *testing.T) {
	expected := "Error reading config file: open invalid/file/path: no such file or directory\n"
	conf, err := load("invalid/file/path")
	if conf != nil || err == nil {
		t.Fatal("Should return error and nil pointer")
	}
	if err.Error() != expected {
		t.Fatalf("got: '%s'; expected: '%s'", err.Error(), expected)
	}
}

func TestLoadInvalidYaml(t *testing.T) {
	conf, err := load("testdata/invalid-yaml.txt")
	if conf != nil || err == nil {
		t.Fatal("Should return error and nil pointer")
	}
}
