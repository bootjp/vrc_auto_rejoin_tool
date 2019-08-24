// +build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func command(instance Instance) *exec.Cmd {
	return &exec.Cmd{
		Path:  os.Getenv("COMSPEC"),
		Stdin: os.Stdin,
		SysProcAttr: &syscall.SysProcAttr{
			CmdLine: `/S /C start vrchat://launch?id=` + instance.ID,
		},
	}
}
