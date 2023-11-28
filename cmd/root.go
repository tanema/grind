package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tanema/grind/lib/procfile"
	"github.com/tanema/grind/lib/runner"
	"github.com/tanema/grind/lib/term"
)

var (
	pfile *procfile.Procfile

	rootCmd = &cobra.Command{
		Version: "0.0.1",
		Use:     "grind",
		Long: `Get on your grind 👑
Run all of your services concurrently within a nix-shell`,
		PersistentPreRun: ensureNix,
	}
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new grind.yml file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return procfile.Create()
		},
	}
	runCmd = &cobra.Command{
		Use:          "run",
		Short:        "Run all services in their own nix-shell.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.New(pfile).RunServices(args)
		},
	}
	shellCmd = &cobra.Command{
		Use:   "shell [service]",
		Short: "Start up interactive shell with deps.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.New(pfile).RunShell(args[0])
		},
	}
	execCmd = &cobra.Command{
		Use:   "exec [service] -- [arbitrary shell commands]",
		Short: "run a command within the environment",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.New(pfile).RunCommand(args[0], strings.Join(args[1:], " "))
		},
	}
	envCmd = &cobra.Command{
		Use:    "env [service]",
		Short:  "Output environment variables for different services.",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			var env []string
			if svc := pfile.Services[args[0]]; svc != nil {
				env = svc.Environ()
			} else if task := pfile.Tasks[args[0]]; task != nil {
				env = task.Environ()
			} else {
				fmt.Printf("Uknown service or task %v", args[0])
				return
			}
			fmt.Println(strings.Join(env, "\n"))
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
	var err error
	var file string

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetUsageFunc(usage)
	rootCmd.SetHelpFunc(help)
	rootCmd.PersistentFlags().StringVarP(&file, "file", "f", "./grind.yml", "Specify a grindfile path to load.")

	pfile, err = procfile.Parse(file)
	if !os.IsNotExist(err) {
		cobra.CheckErr(err)
	} else if pfile == nil {
		rootCmd.AddCommand(initCmd)
		return
	}
	rootCmd.AddCommand(runCmd, envCmd, shellCmd, execCmd)
	for name, task := range pfile.Tasks {
		use := name
		if task.Usage != "" {
			use = task.Usage
		}
		rootCmd.AddCommand(&cobra.Command{
			Use:    use,
			Hidden: task.Hidden,
			Short:  task.Description,
			RunE:   runTask(name),
		})
	}
}

func runTask(taskName string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return runner.New(pfile).RunTask(taskName, args)
	}
}

func ensureNix(cmd *cobra.Command, args []string) {
	if _, err := exec.LookPath("nix"); err == nil {
		return
	}
	term.Println(`{{"nix" | cyan | bold}} not found on your system. Please run the following command to install it.

{{"sh <(curl -L https://nixos.org/nix/install) --daemon" | bold}}

Visit {{"https://nixos.org/download.html" | bold | blue}} to learn more about installing nixos.
`, nil)
	os.Exit(1)
}
