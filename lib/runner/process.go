package runner

import (
	"os"

	"github.com/tanema/grind/lib/procfile"
)

// Process captures a single running process
type Process struct {
	runner *Runner
	defn   *procfile.Service
	stdOut *Logger
	stdErr *Logger
}

func newProc(run *Runner, service *procfile.Service, prefix string) *Process {
	return &Process{
		runner: run,
		defn:   service,
		stdOut: &Logger{prefix: prefix, writer: os.Stdout},
		stdErr: &Logger{prefix: prefix, writer: os.Stderr},
	}
}

// Run will start a single process with before and after hooks for it.
func (proc *Process) run() error {
	defer func() {
		for _, cmd := range proc.defn.After {
			if err := proc.exec(cmd); err != nil {
				return
			}
		}
	}()
	for _, cmd := range proc.defn.Before {
		if err := proc.exec(cmd); err != nil {
			return err
		}
	}
	for _, cmd := range proc.defn.Cmd {
		if err := proc.exec(cmd); err != nil {
			return err
		}
	}
	return nil
}

// runCmd will run a command with the ability to gracefully stop it.
func (proc *Process) exec(cmd string) error {
	proc.stdOut.Printf("🚀 => %v\n", cmd)
	keep, err := proc.defn.EnvKeys()
	if err != nil {
		return err
	}
	env, err := proc.defn.Environ()
	if err != nil {
		return err
	}
	cmdProc := proc.runner.WithShell(cmd, keep...)
	cmdProc.Dir = proc.defn.Dir
	cmdProc.Stdout = proc.stdOut
	cmdProc.Stderr = proc.stdErr
	cmdProc.Env = env
	if err := cmdProc.Start(); err != nil {
		return err
	}
	running[cmdProc.Process.Pid] = cmdProc.Process
	defer func(pid int) { delete(running, pid) }(cmdProc.Process.Pid)
	return cmdProc.Wait()
}
