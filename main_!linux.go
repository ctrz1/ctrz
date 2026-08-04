//go:build !linux

package main

import (
	"log"
)

func main() {
	log.Fatal("ctrz is currently only available for linux")
}
