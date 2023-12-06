# ⚙️  Grind ⚙️

*Fast, reproducible, and concurrent development environment and process manager*

`grind` incorporates ideas from procfiles, makefiles, and uses the nix package
manager to create a development environment that is reproducible and lightweight.
It was created out of a desire for something with the reproducibility of docker
without the resources. `nix` satisfies a lot of this however it is an incredibly
complex tool that is not always easy to setup. It also does not provide an easy
way to run multiple processes at the same time, like foreman.

- [Requirements](#requirements)
- [Usage](#usage)
- [grind.yml Spec](/docs/grind_spec.md)
- [Running Envs](/docs/running_env.md)
- [FAQ](#faq)

### Requirements
`grind` requires a few things to run properly.

- Linux or MacOS (Windows is not supported by nix)
- [nix package manager](https://nixos.org/download.html) `sh <(curl -L https://nixos.org/nix/install) --daemon`

### Usage
To get started quickly, run `grind init` in your project. Once your project has 
a `grind.yml` file in it, you will need to define services and the dependencies 
that they require. You can find out more on how to define these in the 
[grind.yml Spec documentation](/docs/grind_spec.md). Once you have defined these, 
simply run `grind run` to concurrently run any services that you have defined 
in a `nix-shell`. Any tasks that are defined will be outputted in the help usage 
as well and can be run with `grind [task-name]`. Run `grind help` to see the 
detailed output.

### FAQ

- *Why grind*: `grind` stands for *GR*ind *I*s *N*ot *D*ocker. Named so because
  I wanted a tool that was a lot more light weight for development, and did not
  kill my battery.
- *Why tho?*: Because I also wanted to automate some things I use nix for and a
  few other tools all into one.
- *Is this true nix?*: No honestly it is a misuse of nix really. I know this is 
  not how you're supposed to use it, but I find this a lot more usable and 
  straight forward.
