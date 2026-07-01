//go:build !linux
// +build !linux

package main

import "fmt"

func main() {
	fmt.Println("ctrz is currently only available for linux")
}
