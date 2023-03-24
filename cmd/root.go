package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanema/gluey"

	"github.com/tanema/grind/lib/envfile"
	"github.com/tanema/grind/lib/nix"
	"github.com/tanema/grind/lib/procfile"
	"github.com/tanema/grind/lib/runner"
)

type Flags struct {
	Dir    string
	File   string
	Env    []string
	Only   []string
	Except []string
}

var (
	env   *envfile.Env
	pfile *procfile.Procfile
	deps  *nix.Nix
	flags = Flags{
		File: "grind.yml",
	}

	rootCmd = &cobra.Command{
		Version: "0.0.1",
		Use:     "grind",
		Long: `Fast, reproducible, development environment and process manager

- Run multiple processes with a single command.
- Easily manage dependencies with nix.
- Defined custom tasks in the context of a service.
- Run an isolated shell with prerequisites satisfied.

Get on your grind 👑`,
	}
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new grind.yml file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return procfile.Create()
		},
	}
	runCmd = &cobra.Command{
		Use:     "run",
		Short:   "Ensure dependencies are satisfied and start up all specified services.",
		PreRunE: resolveDeps,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.New(deps, pfile).RunServices(flags.Only, flags.Except)
		},
	}
	shellCmd = &cobra.Command{
		Use:     "shell",
		Short:   "Start up interactive shell with deps.",
		PreRunE: resolveDeps,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.New(deps, pfile).RunShell()
		},
	}
	execCmd = &cobra.Command{
		Use:     "exec -- [arbitrary shell commands]",
		Short:   "run a command within the environment",
		PreRunE: resolveDeps,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.New(deps, pfile).RunCommand(strings.Join(args, " "))
		},
	}
	envCmd = &cobra.Command{
		Use:     "env",
		Short:   "Output environment variables for different services.",
		PreRunE: resolveDeps,
		Hidden:  true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				env, _ := pfile.Environ()
				fmt.Println(strings.Join(env, "\n"))
				return
			}
			name := args[0]
			if svc := pfile.Services[name]; svc != nil {
				env, _ := svc.Environ()
				fmt.Println(strings.Join(env, "\n"))
			} else if task := pfile.Tasks[name]; task != nil {
				env, _ := task.Environ()
				fmt.Println(strings.Join(env, "\n"))
			} else {
				fmt.Printf("Uknown service or task %v", name)
			}
		},
	}
)

// Execute is the main app entrypoint
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetUsageFunc(usage)
	rootCmd.SetHelpFunc(help)

	pwd, err := os.Getwd()
	cobra.CheckErr(err)
	rootCmd.PersistentFlags().StringVarP(&flags.Dir, "root", "d", pwd, "Specify an alternate application root.")
	rootCmd.PersistentFlags().StringVarP(&flags.File, "file", "f", "grind.yml", "Specify an alternate Procfile to load, implies -d as the root.")
	rootCmd.PersistentFlags().StringSliceVarP(&flags.Env, "env", "e", nil, "Specify one or more .env files to load")

	runCmd.Flags().StringSliceVarP(&flags.Only, "only", "o", nil, "Specify one or more services to run: --only server,db.")
	runCmd.Flags().StringSliceVarP(&flags.Except, "except", "x", nil, "Specify one or more services to exclude from run: --except server,db.")

	env, err = envfile.Parse(flags.Env...)
	cobra.CheckErr(err)

	pfile, err = procfile.Parse(flags.Dir, flags.File, env)
	if !os.IsNotExist(err) {
		cobra.CheckErr(err)
	}

	if pfile != nil {
		rootCmd.AddCommand(runCmd)
		rootCmd.AddCommand(envCmd)

		if len(pfile.Nixpkgs) > 0 {
			rootCmd.AddCommand(shellCmd)
			rootCmd.AddCommand(execCmd)
		}

		for name, task := range pfile.Tasks {
			addTask(name, task.Description)
		}
	} else {
		rootCmd.AddCommand(initCmd)
	}
}

func addTask(name, desc string) {
	rootCmd.AddCommand(&cobra.Command{
		Use:     name,
		Short:   desc,
		PreRunE: resolveDeps,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.New(deps, pfile).RunTask(name)
		},
	})
}

func resolveDeps(cmd *cobra.Command, args []string) error {
	rnd := gluey.New()

	deps = nix.New(pfile)
	if len(pfile.Nixpkgs) == 0 {
		return nil
	}

	err := rnd.InFrame("📦 Dependencies", func(c *gluey.Ctx, f *gluey.Frame) error {
		spinGrp := c.NewSpinGroup()
		for _, pkg := range pfile.Nixpkgs {
			func(pkg string) {
				spinGrp.Go(pkg, func(spinner *gluey.Spinner) error {
					return deps.Install(pkg)
				})
			}(pkg)
		}
		errs := spinGrp.Wait()
		rnd.Debreif(errs)
		if len(errs) > 0 {
			return fmt.Errorf("could not install all requested packages")
		}
		return nil
	})

	if err != nil {
		fmt.Printf(`Error installing packages. Please try searching to ensure you have the correct package name:

https://search.nixos.org/packages
`)
	}

	return err
}
