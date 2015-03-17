// +build windows

package spy

import (
	"os/exec"
	"strconv"
	"syscall"
)

func setProcessGroupID(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func killByProcessGroupID(cmd *exec.Cmd) error {
	return exec.Command("taskkill", "/t", "/pid", strconv.Itoa(cmd.Process.Pid)).Run()
}
