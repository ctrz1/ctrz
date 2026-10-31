package config

import (
	"ctrz/spec"
	"fmt"
	"testing"
)

func TestLoader(t *testing.T) {
	expected := spec.ContainerSpec{
		Name:        "demo",
		Ports:       []string{"8443:8443"},
		CPU:         "80",
		MemoryLimit: 15,
		Detach:      true,
		Command:     []string{"demo-webserver"},
		Environment: []string{"KEY=VALUE"},
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
	files := []string{
		"testdata/invalid-yaml.txt",
		"testdata/invalid-yaml2.txt",
	}
	for _, file := range files {
		conf, err := load(file)
		if conf != nil || err == nil {
			fmt.Printf("Expected empty conf, got %+v\n", *conf)
			t.Fatal("Should return error and nil pointer")
		}
	}
}
