package nix

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/tanema/grind/lib/procfile"
)

type (
	// DependencyStatus describes the installation status
	DependencyStatus string
	// Nix is the dependency manager using nixos
	Nix struct {
		pfile *procfile.Procfile
		Deps  []*Dependency
	}
	// Dependency is a single dep that is required
	Dependency struct {
		Require *procfile.Requirment
		Status  DependencyStatus
		Pkg     nixPkg
		Error   exec.Error
	}
)

const (
	// Installed means the dep is satisfied
	Installed DependencyStatus = "installed"
	// ArbitraryVersion means the dep is satisfied however there are many versions and the build may not be reproducable
	ArbitraryVersion DependencyStatus = "arbitrary version"
	// NotInstalled means that the dep is missing
	NotInstalled DependencyStatus = "not installed"
	// Unknown means that the process stopped before figuring out the status
	Unknown DependencyStatus = "unknown"
)

// New will create a new dependency manager
func New(pfile *procfile.Procfile) *Nix {
	deps := []*Dependency{}
	for _, req := range pfile.Requires {
		deps = append(deps, &Dependency{
			Require: req,
			Status:  Unknown,
		})
	}

	var wg sync.WaitGroup
	for _, dep := range deps {
		wg.Add(1)
		go func(dep *Dependency) {
			defer wg.Done()
			dep.query()
		}(dep)
	}
	wg.Wait()

	return &Nix{pfile: pfile, Deps: deps}
}

// AllSatisfied will return true if all deps are installed
func (nix *Nix) AllSatisfied() bool {
	for _, dep := range nix.Deps {
		if dep.Status == NotInstalled || dep.Status == Unknown {
			return false
		}
	}
	return true
}

// AllPinned will return true if all deps are pinned to an attr name
func (nix *Nix) AllPinned() bool {
	for _, dep := range nix.Deps {
		if dep.Require.Attr == "" {
			return false
		}
	}
	return true
}

// Resolve will install any missing packages
func (dep *Dependency) Resolve() error {
	if dep.Status != NotInstalled {
		return nil
	}
	args := []string{
		"--install",
		"--attr",
		fmt.Sprintf("nixpkgs.%v", dep.Require.Name),
	}
	_, err := exec.Command("nix-env", args...).Output()
	if err != nil {
		if execErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf(string(execErr.Stderr))
		}
		return err
	}
	return dep.query()
}

// WithShell will create a new command with nix-shell satisfying the deps
func (nix *Nix) WithShell(ctx context.Context, cmd string) *exec.Cmd {
	args := []string{"<nixpkgs>"}
	if nix.pfile.Isolated {
		args = append(args, "--pure")
	}
	args = append(args, "-p")
	for _, pkg := range nix.Deps {
		args = append(args, pkg.Require.Attr)
	}
	args = append(args, "--command", cmd)
	return exec.CommandContext(ctx, "nix-shell", args...)
}

// WithInteractiveShell will create a new command with nix-shell satisfying the deps
func (nix *Nix) WithInteractiveShell(ctx context.Context) *exec.Cmd {
	args := []string{"<nixpkgs>"}
	if nix.pfile.Isolated {
		args = append(args, "--pure")
	}
	args = append(args, "-p")
	for _, pkg := range nix.Deps {
		args = append(args, pkg.Require.Attr)
	}
	return exec.CommandContext(ctx, "nix-shell", args...)
}
