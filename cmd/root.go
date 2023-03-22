package cmd

import (
	"os"

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
	flags = Flags{
		File: "grind.yml",
	}

	rootCmd = &cobra.Command{
		Use: "grind",
		Long: `Fast, reproducible, development environment and process manager

grind incororates ideas from procfiles, makefiles, and uses the nix package
manager to create a development environment that is reproducible and lightweight.
It was created out of a desire for something with the reproducibility of docker
without the resources. nix satisfies a lot of this however it is an incredibly
complex tool that is not always easy to setup. It also does not provide an easy
way to run multiple processes at the same time, like foreman.

Get on your grind, royalty 👑`,
	}

	run = &cobra.Command{
		Use:   "run",
		Short: "Ensure dependencies are satisfied and start up all specified services.",
		Run:   runCmd,
	}
	shell = &cobra.Command{
		Use:   "shell",
		Short: "Start up interactive shell with deps.",
		Run:   shellCmd,
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

	pwd, err := os.Getwd()
	cobra.CheckErr(err)
	rootCmd.PersistentFlags().StringVarP(&flags.Dir, "root", "d", pwd, "Specify an alternate application root.")
	rootCmd.PersistentFlags().StringVarP(&flags.File, "file", "f", "grind.yml", "Specify an alternate Procfile to load, implies -d as the root.")
	rootCmd.PersistentFlags().StringSliceVarP(&flags.Env, "env", "e", nil, "Specify one or more .env files to load")

	run.Flags().StringSliceVarP(&flags.Only, "only", "o", nil, "Specify one or more services to run: --only server,db.")
	run.Flags().StringSliceVarP(&flags.Except, "except", "x", nil, "Specify one or more services to exclude from run: --except server,db.")
	rootCmd.AddCommand(run)
	rootCmd.AddCommand(shell)

	env, err = envfile.Parse(flags.Env...)
	cobra.CheckErr(err)

	pfile, err = procfile.Parse(flags.Dir, flags.File, env)
	cobra.CheckErr(err)

	if pfile != nil {
		for name, task := range pfile.Tasks {
			rootCmd.AddCommand(&cobra.Command{
				Use:   name,
				Short: task.Description,
				Run:   runTaskCmd,
			})
		}
	}
}

func runCmd(cmd *cobra.Command, args []string) {
	deps, err := resolveDeps(pfile)
	cobra.CheckErr(err)
	runner.New(deps, runner.Config{
		Procfile: pfile,
		Only:     flags.Only,
		Except:   flags.Except,
	}).RunServices()
}

func shellCmd(cmd *cobra.Command, args []string) {
	deps, err := resolveDeps(pfile)
	cobra.CheckErr(err)
	runner.New(deps, runner.Config{Procfile: pfile}).RunShell()
}

func runTaskCmd(cmd *cobra.Command, args []string) {
	deps, err := resolveDeps(pfile)
	cobra.CheckErr(err)
	runner.New(deps, runner.Config{Procfile: pfile}).RunTask(cmd.Use)
}

func resolveDeps(pfile *procfile.Procfile) (*nix.Nix, error) {
	rnd := gluey.New()
	deps := nix.New(pfile)

	if !deps.AllSatisfied() {
		err := rnd.InFrame("📦 Installing Dependencies", func(c *gluey.Ctx, f *gluey.Frame) error {
			spinGrp := c.NewSpinGroup()
			for _, dep := range deps.Deps {
				func(dep *nix.Dependency) {
					spinGrp.Go(dep.Require.Name, func(spinner *gluey.Spinner) error {
						return dep.Resolve()
					})
				}(dep)
			}
			rnd.Debreif(spinGrp.Wait())
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	if !deps.AllPinned() {
		if err := rnd.Spinner("📌 Pinning versions", func(s *gluey.Spinner) error {
			return deps.Setup()
		}); err != nil {
			return nil, err
		}
	}

	return deps, nil
}
