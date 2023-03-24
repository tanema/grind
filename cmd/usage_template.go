package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tanema/grind/lib/term"
)

const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}
`

const usageTemplate = `{{"Usage:" | bold | bright}} {{"grind [command]" | cyan }}
{{- if .HasAvailableSubCommands}}

{{"Available Commands:" | bold | bright}}
{{- range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding | bold }} {{.Short}}{{end}}{{end}}
{{- end}}
{{- if .HasAvailableLocalFlags}}

{{"Flags:" | bold | bright}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}
{{- if .HasAvailableInheritedFlags}}

{{"Global Flags:" | bold | bright}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}
{{- if .HasAvailableSubCommands}}

Use {{"grind [command] --help" | bold | bright}} for more information about a command.
{{- end}}
`

func usage(c *cobra.Command) error {
	return term.Println(usageTemplate, c)
}

func help(c *cobra.Command, args []string) {
	term.Println(helpTemplate, c)
	term.Println(usageTemplate, c)
}
