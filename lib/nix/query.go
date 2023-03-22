package nix

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type (
	nixPkg struct {
		Name       string         `json:"name"`
		OutputName string         `json:"outputName"`
		Outputs    map[string]any `json:"outputs"`
		PName      string         `json:"pname"`
		System     string         `json:"system"`
		Version    string         `json:"version"`
		Meta       nixPkgMeta     `json:"meta"`
	}
	nixPkgMeta struct {
		Available   bool     `json:"available"`
		Broken      bool     `json:"broken"`
		Description string   `json:"description"`
		Homepage    string   `json:"homepage"`
		Insecure    bool     `json:"insecure"`
		Name        string   `json:"name"`
		Platforms   []string `json:"platforms"`
		Position    string   `json:"position"`
		Unfree      bool     `json:"unfree"`
		Unsupported bool     `json:"unsupported"`
	}
)

func (dep *Dependency) query() error {
	if dep.Require.Attr != "" {
		dep.Status = Installed
		return nil
	}

	available := map[string]nixPkg{}

	args := []string{
		"--file",
		"'<nixpkgs>'",
		"--query",
		dep.Require.Name,
		"--json",
		"--meta",
	}

	output, err := exec.Command("nix-env", args...).Output()
	if err != nil {
		dep.Status = NotInstalled
		return nil
	} else if err := json.Unmarshal(output, &available); err != nil {
		return err
	}

	packages := make([]nixPkg, 0, len(available))
	for _, pkg := range available {
		packages = append(packages, pkg)
	}

	if len(packages) > 1 {
		dep.Status = ArbitraryVersion
	} else if len(packages) == 1 {
		dep.Status = Installed
	}

	dep.Pkg = packages[0]
	return nil
}

// Setup will ensure we have all the attr names to use in shell
func (nix *Nix) Setup() error {
	args := []string{
		"--json",
		"--file",
		"<nixpkgs>",
		"--query",
		"--available",
		"--attr-path",
	}
	for _, pkg := range nix.Deps {
		if pkg.Require.Attr == "" {
			args = append(args, pkg.Require.Name)
		}
	}

	output, err := exec.Command("nix-env", args...).Output()

	if err != nil {
		if execErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf(string(execErr.Stderr))
		}
		return err
	}

	if err != nil {
		return err
	}

	allPkgs := map[string]nixPkg{}
	if err := json.Unmarshal(output, &allPkgs); err != nil {
		return err
	}

	for attrName, pkg := range allPkgs {
		for _, dep := range nix.Deps {
			if dep.Pkg.Name == pkg.Name {
				dep.Require.Attr = attrName
				break
			}
		}
	}

	return nix.pfile.Save()
}
