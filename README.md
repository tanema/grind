# ⚙️  Grind ⚙️  [Work In Progress]

*Fast, reproducible, and concurrent development environment and process manager*

`grind` incororates ideas from procfiles, makefiles, and uses the nix package
manager to create a development environment that is reproducible and lightweight.
It was created out of a desire for something with the reproducibility of docker
without the resources. `nix` satisfies a lot of this however it is an incredibly
complex tool that is not always easy to setup. It also does not provide an easy
way to run multiple processes at the same time, like foreman.

- [Requirements](#requirements)
- [Usage](#usage)
- [Demos](#demos)
- [grind.yml Spec](#grindyml-spec-definition)
- [FAQ](#faq)

### Requirements
`grind` requires a few things to run properly.

- [nix package manager](https://nixos.org/download.html)
- Linux or MacOS (Windows is not supported by nix)

### Usage
Once your project has a `grind.yml` file in it, simply run `grind run` to resolve
dependencies and run the projects. Any tasks that are defined will be outputted
in the help usage as well.


```
Usage:
  grind [command]

Available Commands:
  exec        run a command within the environment
  help        Help about any command
  run         Ensure dependencies are satisfied and start up all specified services.
  shell       Start up interactive shell with deps.
  ...         Any defined tasks will show in the help description

Flags:
  -e, --env strings   Specify one or more .env files to load
  -f, --file string   Specify an alternate Procfile to load, implies -d as the root. (default "grind.yml")
  -h, --help          help for grind
  -d, --root string   Specify an alternate application root. (default "/Users/timanema/workspace/tubes.dev")
```

### Demos
Using an isolated version of Go for this project only:

https://user-images.githubusercontent.com/463193/227263067-f5130080-f39d-44a9-b9a4-bb642a92fc00.mov

Using `grind` to run multiple services with many dependencies:

https://user-images.githubusercontent.com/463193/227267280-5a168169-7321-4572-9980-59716d6e332a.mov

### grind.yml Spec Definition.
Each grind project contains a grind.yaml file to specify requirments, services,
and tasks. An spec for a project that has a go backend and a webpack built  front-end
would look something like this.

```yaml
version: '1'

# Set if shell, services and tasks are run in an isolated shell with no
# outside dependencies or OS Env vars set within the running process.
# This is ideal for being able to improve reproducability but some tools might
# not work in isolation.
isolated: false

# Set environment variables for all services and tasks.
env:
  ENV: development
  PORT: 8080

# Dependency requirments, satisfied by nix.
# Search for packages with https://search.nixos.org/packages
nixpkgs:
  - nodejs-18_x
  - go_1_20

# Services are the services run when running the application with `grind run`
services:
  client:
    dir: ./client
    env:
      DEBUG: 1
    before:
      - npm i
    cmds:
      - npx webpack -w
    after:
      - rm -rf ./dist
  server:
    dir: ./server
    env:
      PORT: 8080
    before:
      - go mod tidy
    cmds:
      - go run main.go
    after:
      - echo "shutdown"

# tasks are makefile like tasks, run in each service's context
tasks:
  deploy: # defined `grind deploy`
    service: server
    cmds:
      - gcloud app deploy app.yml --project=my-project1a
  test:
    service: server
    cmds:
      - go test ./...
```

### FAQ

- *Why grind*: `grind` stands for *GR*ind *I*s *N*ot *D*ocker. Named so because
  I wanted a tool that was a lot more light weight for development, and did not
  kill my battery.
- *Why tho?*: Because I also wanted to automate some things I use nix for and a
  few other tools all into one.
