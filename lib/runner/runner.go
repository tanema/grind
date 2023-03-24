package runner

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/fatih/color"
	"golang.org/x/exp/slices"

	"github.com/tanema/grind/lib/nix"
	"github.com/tanema/grind/lib/procfile"
)

var logColors = []*color.Color{
	color.New(color.FgHiYellow),
	color.New(color.FgHiBlue),
	color.New(color.FgHiMagenta),
	color.New(color.FgHiCyan),
	color.New(color.FgHiWhite),
	color.New(color.FgYellow),
	color.New(color.FgBlue),
	color.New(color.FgMagenta),
	color.New(color.FgCyan),
}

var running = map[int]*os.Process{}

type (
	// Config is configuration for the runner
	Config struct {
		Procfile *procfile.Procfile
		Only     []string
		Except   []string
	}
	// Runner coordinates between many processes
	Runner struct {
		procfile  *procfile.Procfile
		ctx       context.Context
		cancel    context.CancelFunc
		nix       *nix.Nix
		processes map[string]*Process
		tasks     map[string]*Process
		sigc      chan os.Signal
	}
)

// New creates a new runner for a parsed procfile
func New(deps *nix.Nix, pfile *procfile.Procfile) *Runner {
	ctx, cancel := context.WithCancel(context.Background())

	maxTitleLen := 0
	for procName := range pfile.Services {
		if maxTitleLen < len(procName) {
			maxTitleLen = len(procName)
		}
	}

	runner := &Runner{
		ctx:       ctx,
		cancel:    cancel,
		nix:       deps,
		procfile:  pfile,
		processes: map[string]*Process{},
		tasks:     map[string]*Process{},
	}

	colorIndex := 0
	for name, process := range pfile.Services {
		prefix := logColors[colorIndex].Sprintf("%*v | ", maxTitleLen, name)
		runner.processes[name] = newProc(runner, process, prefix)
		colorIndex = (colorIndex + 1) % len(logColors)
	}

	for name, task := range pfile.Tasks {
		prefix := logColors[colorIndex].Sprintf("%*v | ", maxTitleLen, name)
		runner.tasks[name] = newProc(runner, task, prefix)
		colorIndex = (colorIndex + 1) % len(logColors)
	}
	return runner
}

// RunServices will start all of the default services
func (runner *Runner) RunServices(only, except []string) error {
	go runner.monitorSignals()
	var wg sync.WaitGroup
	var errors = []error{}
	for name, proc := range runner.processes {
		if (only != nil && !slices.Contains(only, name)) || slices.Contains(except, name) {
			continue
		}

		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			if err := proc.run(); err != nil {
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
func (runner *Runner) RunTask(name string) error {
	go runner.monitorSignals()
	return runner.tasks[name].run()
}

// WithShell will run a command in a nix-shell if there are deps and in the
// current shell if there are no deps
func (runner *Runner) WithShell(cmd string, keep ...string) *exec.Cmd {
	if len(runner.procfile.Nixpkgs) > 0 {
		return runner.nix.WithShell(runner.ctx, cmd, keep...)
	}
	parts := strings.Split(cmd, " ")
	return exec.CommandContext(runner.ctx, parts[0], parts[1:]...)
}

// RunShell will start an interactive shell with deps
func (runner *Runner) RunShell() error {
	keep, err := runner.procfile.EnvKeys()
	if err != nil {
		return err
	}
	cmdProc := runner.nix.WithInteractiveShell(runner.ctx, keep...)
	cmdProc.Stdin = os.Stdin
	cmdProc.Stdout = os.Stdout
	cmdProc.Stderr = os.Stderr
	env, err := runner.procfile.Environ()
	if err != nil {
		return err
	}
	cmdProc.Env = env
	return cmdProc.Run()
}

// RunCommand will run a command within the nix-shell
func (runner *Runner) RunCommand(cmd string) error {
	keep, err := runner.procfile.EnvKeys()
	if err != nil {
		return err
	}
	cmdProc := runner.WithShell(cmd, keep...)
	cmdProc.Stdin = os.Stdin
	cmdProc.Stdout = os.Stdout
	cmdProc.Stderr = os.Stderr
	env, err := runner.procfile.Environ()
	if err != nil {
		return err
	}
	cmdProc.Env = env
	return cmdProc.Run()
}

func (runner *Runner) monitorSignals() {
	runner.sigc = make(chan os.Signal, 1)
	signal.Notify(runner.sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	runner.killAll(<-runner.sigc)
}

func (runner *Runner) killAll(sig os.Signal) {
	numProcs := len(running)
	color.Cyan("Shutting down %v processes...", numProcs)
	for _, proc := range running {
		runner.killProc(proc, sig)
	}
	runner.cancel()
	color.Green("Successfully shutdown %v processes", numProcs)
}

func (runner *Runner) killProc(proc *os.Process, sig os.Signal) {
	defer func() { delete(running, proc.Pid) }()
	if err := proc.Signal(sig); err != nil {
		if err := syscall.Kill(proc.Pid, syscall.SIGKILL); err != nil {
			proc.Kill()
		}
	}
}
