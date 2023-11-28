package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tanema/grind/lib/procfile"
	"github.com/tanema/grind/lib/term"
)

const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}
`

const usageTemplate = `{{"Usage:" | bold | bright}}{{if .Cmd.Runnable}}
  {{.Cmd.UseLine | cyan}}{{end}}{{if .Cmd.HasAvailableSubCommands}}
  {{.Cmd.CommandPath | cyan}} [command]{{end}}

{{- if (and (not .Cmd.HasParent) .Def)}}

{{"Services:" | bold | bright}} run with {{"grind run" | cyan}}
{{- range .Def.Services}}
  {{- if (not .Hidden)}}
  {{rpad .Name 11 | bold | bright}}{{if .Description}}{{.Description}}{{else}}{{"no description" | faint}}{{end}}
  {{- end }}{{ end }}
{{- end }}

{{- if .Cmd.HasAvailableSubCommands}}

{{"Available Commands:" | bold | bright}}
{{- range .Cmd.Commands}}
  {{- if .IsAvailableCommand}}
  {{rpad .Name .NamePadding | bold }} {{.Short}}
	{{- end}}{{end}}
{{- end}}

{{- if .Cmd.HasAvailableLocalFlags}}

{{"Flags:" | bold | bright}}
{{.Cmd.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}

{{- if .Cmd.HasAvailableInheritedFlags}}

{{"Global Flags:" | bold | bright}}
{{.Cmd.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}

{{- if .Cmd.HasAvailableSubCommands}}
{{ if not .Def}}
Run {{"grind init" | cyan}} to start a project
{{- end}}
Use {{"grind [command] --help" | bold | bright}} for more information about a command.
{{- end}}
`

func usage(c *cobra.Command) error {
	return term.Println(usageTemplate, struct {
		Cmd *cobra.Command
		Def *procfile.Procfile
	}{
		Cmd: c,
		Def: pfile,
	})
}

func help(c *cobra.Command, args []string) {
	term.Println(helpTemplate, c)
	if err := usage(c); err != nil {
		panic(err)
	}
}
