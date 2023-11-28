package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/exp/slices"

	"github.com/tanema/grind/lib/procfile"
)

type (
	// Config is configuration for the runner
	Config struct {
		Procfile *procfile.Procfile
		Only     []string
		Except   []string
	}
	// Runner coordinates between many processes
	Runner struct {
		procfile *procfile.Procfile
		running  map[int]*exec.Cmd
		ctx      context.Context
		cancel   context.CancelFunc
		sigc     chan os.Signal
		titleLen int
	}
)

// New creates a new runner for a parsed procfile
func New(pfile *procfile.Procfile) *Runner {
	ctx, cancel := context.WithCancel(context.Background())

	maxTitleLen := 0
	for procName := range pfile.Services {
		if maxTitleLen < len(procName) {
			maxTitleLen = len(procName)
		}
	}

	runner := &Runner{
		ctx:      ctx,
		running:  map[int]*exec.Cmd{},
		cancel:   cancel,
		procfile: pfile,
		titleLen: maxTitleLen,
		sigc:     make(chan os.Signal, 1),
	}

	signal.Notify(runner.sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-runner.sigc
		runner.cancel()
	}()

	return runner
}

// RunServices will start all of the default services
func (runner *Runner) RunServices(names []string) error {
	procs := []*Process{}
	for name, svc := range runner.procfile.Services {
		if len(names) > 0 && !slices.Contains(names, name) {
			continue
		}
		procs = append(procs, newProc(runner, svc))
	}
	if err := runner.spawn(procs, func(proc *Process) error { return proc.before(true, nil) }); err != nil {
		return err
	}
	defer runner.spawn(procs, func(proc *Process) error { return proc.after(true, nil) })
	return runner.spawn(procs, func(proc *Process) error { return proc.cmd(true, nil) })
}

func (runner *Runner) spawn(procs []*Process, fn func(*Process) error) error {
	var wg sync.WaitGroup
	var errors = []error{}
	for _, proc := range procs {
		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			if err := fn(proc); err != nil {
				errors = append(errors, err)
				runner.cancel()
			}
		}(proc)
	}
	wg.Wait()
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

// RunTask will start a single task
func (runner *Runner) RunTask(name string, args []string) error {
	return runner.runTask(name, false, args)
}

func (runner *Runner) runTask(name string, capture bool, args []string) error {
	task, ok := runner.procfile.Tasks[name]
	if !ok {
		return fmt.Errorf("undefined task %v", name)
	}
	return newProc(runner, task).run(capture, args)
}

// RunShell will start an interactive shell with deps
func (runner *Runner) RunShell(name string) error {
	svc, ok := runner.procfile.Services[name]
	if !ok {
		return fmt.Errorf("undefined service %v", name)
	}
	return newProc(runner, svc).shell()
}

// RunCommand will run a command within the nix-shell
func (runner *Runner) RunCommand(name, cmd string) error {
	svc, ok := runner.procfile.Services[name]
	if !ok {
		return fmt.Errorf("undefined service %v", name)
	}
	return newProc(runner, svc).exec(cmd)
}
