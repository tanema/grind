### grind.yml Spec Definition.
Each grind project contains a grind.yaml file to specify requirments, services,
and tasks. An spec for a project that has a go backend and a webpack built  front-end
would look something like this.

```yaml
version: "1" # Spec version in case we change it in the future
envs: # env files to load vars to set on all of the services
  - config/dev.env
env: # Env vars set for every single service globally
  DEBUG: 1
nixpkgs: [] # nixpkgs that are required for all services. This most likely uneeded

services:
  # services define all of the running services that are required for the main 
  # product run so that any developer can run `grind run` and have the entire
  # service setup
  server:
    desc: "Backend Go server" # description about what the service is, output in help
    dir: server # optional directory that this service is run in
    nixpkgs: [go] # nixpkgs to install before running the service
    env: # env vars that are only set for this service
      PORT: 8081
    before: # commands that will run before the service starts
      - echo "starting"
    cmds: # the main commands to run when running the service
      - go run main.go
    after: # any command that will run after the service stops
      - echo "done."

tasks:
  # tasks are commands that are run within the context of the services.
  test:
    cmds:
      # run another task by name, this is helpful to run commands across several
      # services.
      - .@go-test 
  go-test:
    service: server # define which environment to run this task
    hidden: true # hide this command from help output to guide users to use the main test command
    cmds:
      - go test ./...
```

