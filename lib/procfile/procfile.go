package procfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/tanema/grind/lib/envfile"
)

type (
	// Procfile is the type for the procfile definition
	Procfile struct {
		Dir         string                 `yaml:"-"`
		Filepath    string                 `yaml:"-"`
		Perms       os.FileMode            `yaml:"-"`
		Version     string                 `yaml:"version"`
		Isolated    bool                   `yaml:"isolated,omitempty"`
		FlagEnv     *envfile.Env           `yaml:"-"`
		Envfiles    []string               `yaml:"envs,omitempty"`
		Environment map[string]interface{} `yaml:"env,omitempty"`
		Nixpkgs     []string               `yaml:"nixpkgs,omitempty"`
		Services    map[string]*Service    `yaml:"services,omitempty"`
		Tasks       map[string]*Service    `yaml:"tasks,omitempty"`
	}
	// Requirement describes a dependency's version and pinned attribute
	Requirement struct {
		From string `yaml:"from,omitempty"`
		Attr string `yaml:"attr,omitempty"`
	}
	// Service is a single process description
	Service struct {
		procfile    *Procfile         `yaml:"-"`
		Name        string            `yaml:"-"`
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
	Nixpkgs: []string{"nodejs-18_x"},
	Environment: map[string]any{
		"DEBUG": 1,
		"PORT":  8080,
	},
	Services: map[string]*Service{
		"server": {Dir: "server", Cmd: []string{`echo "start server"`}},
		"client": {Dir: "client", Cmd: []string{`npm init`}},
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
func Parse(dir, filename string, env *envfile.Env) (*Procfile, error) {
	procfile := &Procfile{
		Dir:      dir,
		Filepath: filepath.Join(dir, filename),
		FlagEnv:  env,
	}
	info, err := os.Stat(procfile.Filepath)
	if err != nil {
		return nil, err
	}
	procfile.Perms = info.Mode()
	byteData, err := ioutil.ReadFile(procfile.Filepath)
	if err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalStrict(byteData, &procfile); err != nil {
		return nil, err
	}
	for _, svc := range procfile.Services {
		svc.procfile = procfile
	}
	for name, task := range procfile.Tasks {
		task.procfile = procfile
		if task.Service == "" {
			continue
		}
		task.service = procfile.Services[task.Service]
		if task.service == nil {
			return nil, fmt.Errorf("undefined service %v requested in task %v", task.Service, name)
		}
		task.Dir = task.service.Dir
	}
	return procfile, nil
}

// Environ will generate an array of the variables for the procfile for all
// services and inherits the flag args
func (pfile *Procfile) Environ() ([]string, error) {
	env := []string{}
	if !pfile.Isolated {
		env = append(env, os.Environ()...)
	}
	env = append(env, pfile.FlagEnv.ToArray()...)

	fileEnv, err := envfile.Parse(pfile.Envfiles...)
	if err != nil {
		return nil, err
	}
	env = append(env, fileEnv.ToArray()...)

	for key, val := range pfile.Environment {
		env = append(env, fmt.Sprintf("%v=%v", key, val))
	}
	return env, nil
}

// EnvKeys will collect all the env keys that are set in the procfile. This is
// used for isolated shells to tell nix-shell to keep those values
func (pfile *Procfile) EnvKeys() ([]string, error) {
	keys := pfile.FlagEnv.Keys()

	fileEnv, err := envfile.Parse(pfile.Envfiles...)
	if err != nil {
		return nil, err
	}
	keys = append(keys, fileEnv.Keys()...)

	for key := range pfile.Environment {
		keys = append(keys, key)
	}
	return keys, nil
}

// Environ will generate an array of the variables for a single service, inheriting
// from the procfile and flag args. If the service is a task and inherits a service,
// then it will inherit from that service, then procfile, and flag args
func (svc *Service) Environ() ([]string, error) {
	env := []string{}
	if svc.service != nil {
		parentEnv, err := svc.service.Environ()
		if err != nil {
			return nil, err
		}
		env = append(env, parentEnv...)
	} else {
		parentEnv, err := svc.procfile.Environ()
		if err != nil {
			return nil, err
		}
		env = append(env, parentEnv...)
	}

	fileEnv, err := envfile.Parse(svc.Envfiles...)
	if err != nil {
		return nil, err
	}
	env = append(env, fileEnv.ToArray()...)
	for key, val := range svc.Env {
		env = append(env, fmt.Sprintf("%v=%v", key, val))
	}
	return env, nil
}

// EnvKeys will collect all the env keys that are set for the service. This is
// used for isolated shells to tell nix-shell to keep those values
func (svc *Service) EnvKeys() ([]string, error) {
	keys := []string{}
	if svc.service != nil {
		parentEnv, err := svc.service.EnvKeys()
		if err != nil {
			return nil, err
		}
		keys = append(keys, parentEnv...)
	} else {
		parentEnv, err := svc.procfile.EnvKeys()
		if err != nil {
			return nil, err
		}
		keys = append(keys, parentEnv...)
	}
	fileEnv, err := envfile.Parse(svc.Envfiles...)
	if err != nil {
		return nil, err
	}
	keys = append(keys, fileEnv.Keys()...)
	for key := range svc.Env {
		keys = append(keys, key)
	}
	return keys, nil
}
