// +build !windows

package main

import (
	"os/exec"
	"runtime"
)

func command(instance Instance) *exec.Cmd {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", runArgs+instance.ID)
	}

	return exec.Command("xdg-open", runArgs+instance.ID)
}
