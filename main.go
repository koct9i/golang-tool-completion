package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/urfave/cli/v3"

	"github.com/koct9i/golang-tool-completion/completion"
	"github.com/koct9i/golang-tool-completion/gotool"
)

const (
	rootCommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}} {{if .VisibleFlags}}[global options]{{end}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}
{{- if len .Authors}}

AUTHOR{{template "authorsTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandCategoryTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

GLOBAL OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

GLOBAL OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}

DOCUMENTATION:
   {{.Metadata.DocURL}}{{if .Copyright}}

COPYRIGHT:
   {{template "copyrightTemplate" .}}{{end}}
`

	commandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{template "usageTemplate" .}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .VisiblePersistentFlags}}

GLOBAL OPTIONS:{{template "visiblePersistentFlagTemplate" .}}{{end}}

DOCUMENTATION:
   {{.Metadata.DocURL}}
`

	subcommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .Metadata.DocURL}}

DOCUMENTATION:
   {{.Metadata.DocURL}}{{end}}
`
)

func main() {

	cli.RootCommandHelpTemplate = rootCommandHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.SubcommandHelpTemplate = subcommandHelpTemplate

	cli.HelpFlag = nil

	root := &cli.Command{
		Name:      filepath.Base(os.Args[0]),
		Usage:     "Go is a tool for managing Go source code.",
		ArgsUsage: "[arguments]",
		Description: "This wrapper defines commands/flags/args for help/validation/completion, but execution is transparent:\n" +
			"it always runs the system `go` with the original argv.\n",
		Metadata: map[string]any{
			"DocURL": gotool.DocGoCmd,
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "help",
				Aliases: []string{"h"},
				Usage:   "show help",
				Action: func(ctx context.Context, c *cli.Command, b bool) error {
					if c == c.Root() {
						cli.ShowRootCommandHelp(c)
					} else {
						cli.ShowSubcommandHelp(c)
					}
					return cli.Exit("", 0)
				},
			},
		},
		Commands: []*cli.Command{
			gotool.Bug(),
			gotool.Build(),
			gotool.Clean(),
			gotool.Doc(),
			gotool.Env(),
			gotool.Fix(),
			gotool.Fmt(),
			gotool.Generate(),
			gotool.Get(),
			gotool.Help(),
			gotool.Install(),
			gotool.List(),
			gotool.Mod(),
			gotool.Work(),
			gotool.Run(),
			gotool.Telemetry(),
			gotool.Test(),
			gotool.Tool(),
			gotool.Version(),
			gotool.Vet(),
			completion.Completion(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return nil
		},
		HideHelp:       true,
		ExitErrHandler: func(ctx context.Context, c *cli.Command, err error) {},
	}

	err := root.Run(context.Background(), os.Args)
	if err == nil {
		err = execGo(os.Args[1:])
	}
	if err != nil {
		exitCode := 1
		var exitCoder cli.ExitCoder
		if errors.As(err, &exitCoder) {
			exitCode = exitCoder.ExitCode()
		}
		if exitCode != 0 {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(exitCode)
	}
}

func execGo(args []string) error {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return err
	}
	if goPath == os.Args[0] {
		return fmt.Errorf("recursion detected: found myself instead of go tool")
	}
	return unix.Exec(goPath, append([]string{goPath}, args...), os.Environ())
}
