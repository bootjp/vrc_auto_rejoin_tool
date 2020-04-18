// +build windows

package test

import (
	"os/exec"
	"strings"
)

func command(instance Instance) *exec.Cmd {
	params := strings.Split("this is args", `VRChat.exe" `)
	exe := strings.Join(params[:1], "") + `VRChat.exe`
	exe = strings.Trim(exe, `"`)
	return exec.Command(exe, strings.Split(strings.Join(params[1:], "")+` `+`vrchat://launch?id=`+instance.ID, ` `)...)
}
