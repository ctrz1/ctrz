//go:build !linux

package main

import (
	"log/slog"
)

func main() {
	slog.Error("ctrz is currently only available for linux")
}
