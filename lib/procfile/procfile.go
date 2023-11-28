package procfile

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/tanema/grind/lib/envfile"
)

type (
	// Procfile is the type for the procfile definition
	Procfile struct {
		Dir      string              `yaml:"-"`
		Filepath string              `yaml:"-"`
		Version  string              `yaml:"version"`
		Envfiles []string            `yaml:"envs,omitempty"`
		Env      map[string]string   `yaml:"env,omitempty"`
		Nixpkgs  []string            `yaml:"nixpkgs,omitempty"`
		Services map[string]*Service `yaml:"services,omitempty"`
		Tasks    map[string]*Service `yaml:"tasks,omitempty"`
	}
	// Service is a single process description
	Service struct {
		procfile    *Procfile         `yaml:"-"`
		Hidden      bool              `yaml:"hidden,omitempty"`
		Name        string            `yaml:"-"`
		Usage       string            `yaml:"usage,omitempty"`
		Nixpkgs     []string          `yaml:"nixpkgs,omitempty"`
		Isolated    bool              `yaml:"isolated,omitempty"`
		Description string            `yaml:"desc,omitempty"`
		Service     string            `yaml:"service,omitempty"`
		service     *Service          `yaml:"-"`
		Envfiles    []string          `yaml:"envs,omitempty"`
		Dir         string            `yaml:"dir,omitempty"`
		Env         map[string]string `yaml:"env,omitempty"`
		Before      []string          `yaml:"before,omitempty"`
		Cmd         []string          `yaml:"cmds,omitempty"`
		After       []string          `yaml:"after,omitempty"`
	}
)

var templateProcfile = &Procfile{
	Version: "1",
	Env:     map[string]string{"DEBUG": "1"},
	Services: map[string]*Service{
		"server": {Dir: "server", Cmd: []string{`echo "start server"`}, Env: map[string]string{"PORT": "8080"}},
		"client": {Dir: "client", Cmd: []string{`npm init`}, Nixpkgs: []string{"nodejs-18_x"}},
	},
	Tasks: map[string]*Service{
		"test": {Service: "client", Cmd: []string{`npm test`}},
	},
}

// Create will write out a new procfile
func Create() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(pwd, "grind.yml"))
	if err != nil {
		return err
	}

	byteData, err := yaml.Marshal(templateProcfile)
	if err != nil {
		return err
	}
	_, err = file.Write(byteData)
	return err
}

// Parse will read a procfile and format it validly
func Parse(filename string) (*Procfile, error) {
	fullPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	procfile := &Procfile{
		Dir:      filepath.Dir(fullPath),
		Filepath: fullPath,
	}

	if err := procfile.setup(); err != nil {
		return nil, err
	}

	for name, svc := range procfile.Services {
		if err := svc.setup(name, procfile); err != nil {
			return nil, err
		}
	}

	for name, task := range procfile.Tasks {
		if err := task.setup(name, procfile); err != nil {
			return nil, err
		}
	}
	return procfile, nil
}

func (procfile *Procfile) setup() error {
	byteData, err := os.ReadFile(procfile.Filepath)
	if err != nil {
		return err
	}
	if err := yaml.UnmarshalStrict(byteData, &procfile); err != nil {
		return err
	}
	if procfile.Version != "1" {
		return fmt.Errorf("unknown procfile version %v requested", procfile.Version)
	}
	if procfile.Env == nil {
		procfile.Env = map[string]string{}
	}
	return parseEnvFiles(procfile.Env, nil, procfile.Envfiles...)
}

// Environ will generate an array of the variables for a single service, inheriting
// from the procfile and flag args. If the service is a task and inherits a service,
// then it will inherit from that service, then procfile, and flag args
func (svc *Service) Environ() []string {
	env := []string{}
	if svc.service != nil {
		env = append(env, svc.service.Environ()...)
	}
	if !svc.Isolated {
		env = append(env, os.Environ()...)
	}
	for key, val := range svc.Env {
		env = append(env, key+"="+val)
	}
	return env
}

// EnvKeys will collect all the env keys that are set for the service. This is
// used for isolated shells to tell nix-shell to keep those values
func (svc *Service) EnvKeys() []string {
	keys := []string{}
	if svc.service != nil {
		keys = append(keys, svc.service.EnvKeys()...)
	}
	for key := range svc.Env {
		keys = append(keys, key)
	}
	return keys
}

func (svc *Service) setup(name string, procfile *Procfile) error {
	svc.Name = name
	svc.Dir = filepath.Join(procfile.Dir, svc.Dir)
	svc.Nixpkgs = append(svc.Nixpkgs, procfile.Nixpkgs...)
	svc.procfile = procfile
	if svc.Env == nil {
		svc.Env = map[string]string{}
	}
	if err := parseEnvFiles(svc.Env, procfile.Env, svc.Envfiles...); err != nil {
		return err
	}
	if err := svc.inherit(); err != nil {
		return err
	}
	svc.Env["SVC"] = svc.Name
	svc.Env["TASK"] = svc.Name
	svc.Env["PWD"] = svc.Dir
	return nil
}

func (svc *Service) inherit() error {
	if svc.Service == "" {
		return nil
	} else if svc.procfile.Services[svc.Service] == nil {
		return fmt.Errorf("%v tried to inherit %v which does not exist", svc.Name, svc.Service)
	}
	svc.service = svc.procfile.Services[svc.Service]
	svc.Nixpkgs = append(svc.Nixpkgs, svc.service.Nixpkgs...)
	svc.Dir = svc.service.Dir
	for key, val := range svc.service.Env {
		svc.Env[key] = val
	}
	svc.Env["SVC"] = svc.Service
	return nil
}

func parseEnvFiles(env, ext map[string]string, files ...string) error {
	for key, val := range env {
		env[key] = os.Expand(val, func(v string) string {
			if val, ok := env[v]; ok {
				return val
			} else if val, ok := ext[v]; ok {
				return val
			}
			return os.Getenv(v)
		})
	}
	for key, val := range ext {
		env[key] = val
	}
	return envfile.Parse(env, files...)
}
