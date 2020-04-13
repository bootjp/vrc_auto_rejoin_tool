// +build windows

package main

import (
	"os/exec"
)

func command(instance Instance) *exec.Cmd {
	return exec.Command(runArgs + instance.ID)
}
