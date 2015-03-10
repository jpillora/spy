package spy

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

//process is a restartable exec.Command
type process struct {
	s          *Spy
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
func newProcess(s *Spy, args []string, delay time.Duration) (*process, error) {
	if len(args) == 0 {
		return nil, errors.New("No program specified")
	}
	return &process{
		s:     s,
		prog:  args[0],
		args:  args[1:],
		delay: delay,
		ready: make(chan bool, 1),
	}, nil
}

//runs in it's own goroutine
func (p *process) start() {

	p.ready <- true

	for !p.stopped {
		//only run once ready
		<-p.ready

		cmd := exec.Command(p.prog, p.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		setProcessGroupID(cmd)
		if err := cmd.Start(); err != nil {
			p.s.info("Program failed: %s", err)
			time.Sleep(2 * time.Second)
			continue
		}
		//reset killed flag
		p.killed = false
		//start!
		p.s.debug("Start PID:%v '%s %s'", cmd.Process.Pid, p.prog, strings.Join(p.args, " "))
		p.cmd = cmd
		err := cmd.Wait()

		code := 0
		exerr, ok := err.(*exec.ExitError)
		if ok {
			//TODO confirm this is cross-platform...
			if status, ok := exerr.Sys().(syscall.WaitStatus); ok {
				code = status.ExitStatus()
			} else if !exerr.Success() {
				code = 1
			}
		}
		//if spy did not kill the process,
		//convey to the user that their program exited
		if !p.killed {
			p.s.info("Exit %d", code)
		}
		p.s.debug("Stop PID:%v", cmd.Process.Pid)
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
	p.s.info("Restarting...")
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
