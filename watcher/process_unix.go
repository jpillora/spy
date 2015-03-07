// +build !windows

package watcher

import (
	"os/exec"
	"syscall"
)

func setProcessGroupID(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killByProcessGroupID(cmd *exec.Cmd) error {
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return err
	}
	syscall.Kill(-pgid, 15) // note the minus sign
	return nil
}
