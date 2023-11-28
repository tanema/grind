package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/tanema/grind/lib/procfile"
)

// Process captures a single running process
type Process struct {
	runner *Runner
	defn   *procfile.Service
	prefix string
}

var (
	colorIndex = 0
	logColors  = []*color.Color{
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
)

func newProc(run *Runner, service *procfile.Service) *Process {
	colorIndex = (colorIndex + 1) % len(logColors)
	return &Process{
		runner: run,
		defn:   service,
		prefix: logColors[colorIndex].Sprintf("%*v | ", run.titleLen, service.Name),
	}
}

func (proc *Process) command(cmd string, captured bool, args []string) error {
	var shutdownStart time.Time
	nixCmd := []string{`<nixpkgs>`}
	if proc.defn.Isolated {
		nixCmd = append(nixCmd, "--pure")
		for _, key := range proc.defn.EnvKeys() {
			nixCmd = append(nixCmd, "--keep", key)
		}
	}
	nixCmd = append(nixCmd, append([]string{"--packages"}, proc.defn.Nixpkgs...)...)
	if cmd != "" {
		nixCmd = append(nixCmd, "--command", proc.expandEnv(cmd, args))
	}
	cmdProc := exec.CommandContext(proc.runner.ctx, "nix-shell", nixCmd...)
	cmdProc.Dir = proc.defn.Dir
	cmdProc.Stdin = os.Stdin
	cmdProc.SysProcAttr = &syscall.SysProcAttr{Setpgid: captured}
	cmdProc.Env = proc.defn.Environ()
	cmdProc.WaitDelay = time.Minute
	cmdProc.Cancel = func() error {
		if captured {
			fmt.Fprintln(cmdProc.Stdout, color.CyanString("stopping..."))
		}
		shutdownStart = time.Now()
		return syscall.Kill(-cmdProc.Process.Pid, syscall.SIGKILL)
	}
	cmdProc.Stdout = os.Stdout
	cmdProc.Stderr = os.Stderr
	if captured {
		cmdProc.Stdout = &Logger{prefix: proc.prefix, writer: os.Stdout}
		cmdProc.Stderr = &Logger{prefix: proc.prefix, writer: os.Stderr}
		fmt.Fprintf(cmdProc.Stdout, "🚀 => %v\n", cmd)
	}
	err := cmdProc.Run()
	exited := false
	if exitErr, ok := err.(*exec.ExitError); ok {
		status := exitErr.ProcessState.Sys().(syscall.WaitStatus)
		signal := status.Signal()
		if signal == syscall.SIGKILL || signal == syscall.SIGINT {
			err = nil
			exited = true
		}
	}
	if captured {
		if err != nil {
			fmt.Fprintln(cmdProc.Stdout, color.RedString("🔥 exited with error:"), err)
		} else if exited {
			fmt.Fprintf(cmdProc.Stdout, color.GreenString("✅ exited successfully in %v.\n"), time.Now().Sub(shutdownStart))
		} else {
			fmt.Fprintf(cmdProc.Stdout, color.GreenString("✅ completed successfully in %v.\n"), time.Now().Sub(shutdownStart))
		}
	}
	return err
}

func (proc *Process) run(capture bool, args []string) error {
	if err := proc.before(capture, args); err != nil {
		return err
	}
	defer proc.after(capture, args)
	return proc.cmd(capture, args)
}

func (proc *Process) before(capture bool, args []string) error {
	return proc.runlist(proc.defn.Before, args, capture)
}

func (proc *Process) after(capture bool, args []string) error {
	return proc.runlist(proc.defn.After, args, capture)
}

func (proc *Process) cmd(capture bool, args []string) error {
	return proc.runlist(proc.defn.Cmd, args, capture)
}

func (proc *Process) runlist(cmds, args []string, capture bool) error {
	for _, cmd := range cmds {
		if strings.HasPrefix(cmd, ".@") {
			if err := proc.runner.runTask(strings.TrimPrefix(cmd, ".@"), capture, args); err != nil {
				return err
			}
		} else if err := proc.command(cmd, capture, args); err != nil {
			return err
		}
	}
	return nil
}

// runCmd will run a command with the ability to gracefully stop it.
func (proc *Process) exec(cmd string) error {
	return proc.command(cmd, false, nil)
}

func (proc *Process) shell() error {
	return proc.command("", false, nil)
}

func (proc *Process) expandEnv(cmd string, args []string) string {
	cfg := map[string]string{}
	for i, arg := range args {
		cfg[fmt.Sprintf("%v", i+1)] = arg
	}
	return os.Expand(cmd, func(v string) string {
		if val, ok := proc.defn.Env[v]; ok {
			return val
		} else if val, ok := cfg[v]; ok {
			return val
		}
		return os.Getenv(v)
	})
}
