package watcher

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

//process is a restartable exec.Command
type process struct {
	w          *Watcher
	prog       string
	args       []string
	delay      time.Duration
	cmd        *exec.Cmd
	ready      chan bool
	restarting bool
	stopped    bool
	killed     bool
}

//newProcess creates a new process
func newProcess(w *Watcher, args []string, delay time.Duration) (*process, error) {
	if len(args) == 0 {
		return nil, errors.New("No program specified")
	}
	return &process{
		w:     w,
		prog:  args[0],
		args:  args[1:],
		delay: delay,
		ready: make(chan bool, 1),
	}, nil
}

func (p *process) start() {

	p.ready <- true

	for !p.stopped {
		//only run once ready
		<-p.ready

		p.killed = false

		cmd := exec.Command(p.prog, p.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		setProcessGroupID(cmd)
		if err := cmd.Start(); err != nil {
			p.w.info("Program failed: %s", err)
			time.Sleep(2 * time.Second)
			continue
		}

		//failed to start
		p.w.debug("Start #%v '%s %s'", cmd.Process.Pid, p.prog, strings.Join(p.args, " "))
		//start!
		p.cmd = cmd
		err := cmd.Wait()

		success := true
		exerr, ok := err.(*exec.ExitError)
		if ok {
			success = exerr.Success()
		}
		//if watcher did not kill the process,
		//convey that program exited
		if !p.killed {
			var msg string
			if success {
				msg = "success"
			} else {
				msg = "non-zero"
			}
			p.w.info("Exit " + msg)
		}
		p.w.debug("Stop #%v", cmd.Process.Pid)
		p.cmd = nil
	}
}

func (p *process) restart() {
	//restart already queued
	if p.restarting {
		return
	}
	p.restarting = true
	time.Sleep(p.delay)
	p.w.info("Restarting...")
	//kill process
	p.killed = true
	p.kill()
	if len(p.ready) == 0 {
		p.ready <- true
	}
	p.restarting = false
}

func (p *process) stop() {
	p.stopped = true
	if p.cmd != nil {
		p.kill()
	}
}

func (p *process) kill() {
	if p.cmd == nil || p.cmd.Process == nil {
		return
	}
	//kill process group!
	err := killByProcessGroupID(p.cmd)
	if err != nil {
		//process group kill failed, attempt single process kill
		p.cmd.Process.Kill()
	}
}
