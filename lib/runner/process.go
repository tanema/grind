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
func (proc *Process) run() {
	defer func() {
		for _, cmd := range proc.defn.After {
			proc.exec(cmd)
		}
	}()
	for _, cmd := range proc.defn.Before {
		proc.exec(cmd)
	}
	for _, cmd := range proc.defn.Cmd {
		proc.exec(cmd)
	}
}

// runCmd will run a command with the ability to gracefully stop it.
func (proc *Process) exec(cmd string) {
	proc.stdOut.Printf("🚀 => %v\n", cmd)
	env, err := proc.defn.Environ()
	if err != nil {
		proc.stdErr.Printf("🔥 Error: %v\n", err)
		return
	}
	r := proc.runner
	cmdProc := r.nix.WithShell(r.ctx, cmd)
	cmdProc.Dir = proc.defn.Dir
	cmdProc.Stdout = proc.stdOut
	cmdProc.Stderr = proc.stdErr
	cmdProc.Env = env
	if err := cmdProc.Start(); err != nil {
		proc.stdErr.Printf("🔥 Error: %v\n", cmd)
	}
	running[cmdProc.Process.Pid] = cmdProc.Process
	defer func(pid int) { delete(running, pid) }(cmdProc.Process.Pid)
	if err := cmdProc.Wait(); err != nil {
		proc.stdErr.Printf("🔥 Error: %v\n", err)
	}
}
