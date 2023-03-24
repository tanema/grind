package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tanema/grind/lib/procfile"
	"github.com/tanema/grind/lib/term"
)

const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}
`

const usageTemplate = `{{"Usage:" | bold | bright}} {{"grind [command]" | cyan }}

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
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}

{{- if .Cmd.HasAvailableSubCommands}}
{{ with .Def}}
{{- if eq (len .Nixpkgs) 0}}
Add {{"nixpkgs" | bold}} to your {{"grind.yml" | bold}} to enable the {{"exec" | cyan}} and {{"shell" | cyan}} commands.
{{- end }}
{{- else}}
Run {{"grind init" | cyan}} to start a project
{{- end}}
Use {{"grind [command] --help" | bold | bright}} for more information about a command.
{{- end}}
`

type UsageInfo struct {
	Cmd *cobra.Command
	Def *procfile.Procfile
}

func usage(c *cobra.Command) error {
	return term.Println(usageTemplate, UsageInfo{
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
