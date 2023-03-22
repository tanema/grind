package runner

import (
	"context"
	"os"
	"os/signal"
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
func New(deps *nix.Nix, cfg Config) *Runner {
	ctx, cancel := context.WithCancel(context.Background())

	maxTitleLen := 0
	for procName := range cfg.Procfile.Services {
		if maxTitleLen < len(procName) {
			maxTitleLen = len(procName)
		}
	}

	runner := &Runner{
		ctx:       ctx,
		cancel:    cancel,
		nix:       deps,
		procfile:  cfg.Procfile,
		processes: map[string]*Process{},
		tasks:     map[string]*Process{},
	}

	colorIndex := 0
	for name, process := range cfg.Procfile.Services {
		if (cfg.Only != nil && !slices.Contains(cfg.Only, name)) || slices.Contains(cfg.Except, name) {
			continue
		}
		prefix := logColors[colorIndex].Sprintf("%*v | ", maxTitleLen, name)
		runner.processes[name] = newProc(runner, process, prefix)
		colorIndex = (colorIndex + 1) % len(logColors)
	}

	for name, task := range cfg.Procfile.Tasks {
		prefix := logColors[colorIndex].Sprintf("%*v | ", maxTitleLen, name)
		runner.tasks[name] = newProc(runner, task, prefix)
		colorIndex = (colorIndex + 1) % len(logColors)
	}
	return runner
}

// RunServices will start all of the default services
func (runner *Runner) RunServices() {
	go runner.monitorSignals()
	var wg sync.WaitGroup
	for _, proc := range runner.processes {
		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			proc.run()
		}(proc)
	}
	wg.Wait()
}

// RunTask will start a single task
func (runner *Runner) RunTask(name string) {
	go runner.monitorSignals()
	runner.tasks[name].run()
}

// RunShell will start an interactive shell with deps
func (runner *Runner) RunShell() error {
	cmdProc := runner.nix.WithInteractiveShell(runner.ctx)
	cmdProc.Stdin = os.Stdin
	cmdProc.Stdout = os.Stdout
	cmdProc.Stderr = os.Stderr
	return cmdProc.Run()
}

func (runner *Runner) monitorSignals() {
	runner.sigc = make(chan os.Signal, 1)
	signal.Notify(runner.sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-runner.sigc
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
