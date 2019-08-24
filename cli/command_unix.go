// +build !windows

package main

import (
	"os/exec"
	"runtime"
)

func command(instance Instance) *exec.Cmd {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", `vrchat://launch?id=`+instance.ID)
	}

	return exec.Command("xdg-open", `vrchat://launch?id=`+instance.ID)
}
