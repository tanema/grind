package nix

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/tanema/grind/lib/procfile"
)

// Nix is the dependency manager using nixos
type Nix struct {
	pfile *procfile.Procfile
	pkgs  []string
}

// New will create a new dependency manager
func New(pfile *procfile.Procfile) *Nix {
	return &Nix{pfile: pfile}
}

// Install will install and active the package
func (nix *Nix) Install(pkg string) error {
	args := []string{
		"--install",
		"--attr",
		fmt.Sprintf("nixpkgs.%v", pkg),
	}
	_, err := exec.Command("nix-env", args...).Output()
	if err != nil {
		if execErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf(string(execErr.Stderr))
		}
		return err
	}
	nix.pkgs = append(nix.pkgs, pkg)
	return nil
}

// WithShell will create a new command with nix-shell satisfying the deps
func (nix *Nix) WithShell(ctx context.Context, cmd string, keep ...string) *exec.Cmd {
	args := []string{"<nixpkgs>"}
	if nix.pfile.Isolated {
		args = append(args, "--pure", "--keep")
		args = append(args, keep...)
	}
	args = append(args, "-p")
	args = append(args, nix.pkgs...)
	args = append(args, "--command", cmd)
	return exec.CommandContext(ctx, "nix-shell", args...)
}

// WithInteractiveShell will create a new command with nix-shell satisfying the deps
func (nix *Nix) WithInteractiveShell(ctx context.Context, keep ...string) *exec.Cmd {
	args := []string{"<nixpkgs>"}
	if nix.pfile.Isolated {
		args = append(args, "--pure", "--keep")
		args = append(args, keep...)
	}
	args = append(args, "-p")
	args = append(args, nix.pkgs...)
	return exec.CommandContext(ctx, "nix-shell", args...)
}
