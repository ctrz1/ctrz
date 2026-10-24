package mount

import (
	"fmt"
	"strings"
)

func Environment(userEnv []string) ([]string, error) {
	vars := map[string]string{
		"PATH":    "/bin",
		"USER":    "root",
		"LOGNAME": "root",
		"SHELL":   "/bin/sh",
		"HOME":    "/root",
	}

	for _, e := range userEnv {
		key, value, ok := strings.Cut(e, "=")

		if !ok || key == "" || strings.IndexByte(e, '\x00') >= 0 {
			return nil, fmt.Errorf("invalid environment variable: %q", e)
		}
		vars[key] = value
	}

	env := make([]string, 0, len(vars))

	for key, value := range vars {
		env = append(env, key+"="+value)
	}

	return env, nil
}
