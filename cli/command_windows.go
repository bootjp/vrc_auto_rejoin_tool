// +build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func command(instance Instance) *exec.Cmd {
	return &exec.Cmd{
		Stdin: os.Stdin,

		SysProcAttr: &syscall.SysProcAttr{
			CmdLine: `/S /C start ` + runArgs + ` ` + instance.ID,
		},
	}
}
