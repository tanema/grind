# Running Environment
There are different ways that environment variables are handled in services
and tasks which can help the way you run your applications and setup your
behaviour.

## Grind Variables
Some variables are set by grind to help you 

| Variable | Level   | Description |
|----------|---------|-------------|
| `SVC`    | Service | The name of the service that is running |
| `SVC`    | Task    | The name of the inherited service context that the task runs in |
| `TASK`   | Task    | The name of the task that is running |

## Task Args
When running a task using `grind taskName` you are able to provide positional
arguments. These are used in the same way as bash where `$1` is the first arg
passed in and then `$2` and so on. You can be reference all the args using `$@`

## Isolation

Each service has an `isolated` setting that sets the service to run in a very 
strict isolated environment which makes your service a lot more likely to run 
anywhere, but also makes some things a lot harder to setup.

```yaml
services:
  server:
    desc: "Backend Go server" # description about what the service is, output in help
    isolated: true
```

A service that is isolated means that it is run in a `pure` 
[nix-shell](https://nixos.org/manual/nix/stable/command-ref/nix-shell) and it 
will not set any environment variables except those that are set in the config.
This means if your application needs something from a default environment, you 
will need to explicitly set them.
