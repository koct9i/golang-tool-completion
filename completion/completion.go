package completion

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
)

var (
	WithinCompletion bool
	LastCommand      *cli.Command
)

func generateCompletionScript(shell, command, handler string) (scriptPath string, script string, err error) {
	switch shell {
	case "bash":
		scriptPath = fmt.Sprintf("bash-completion/completions/%s", command)
		script = fmt.Sprintf(`__%s_complete_bash() {
  mapfile -t COMPREPLY < <("%[2]s" completion --complete bash -- "${COMP_WORDS[@]:1:COMP_CWORD}")
}
complete -o bashdefault -o default -F __%[1]s_complete_bash %[1]s
`, command, handler)

	case "fish":
		scriptPath = fmt.Sprintf("fish/vendor_completions.d/%s.fish", command)
		script = fmt.Sprintf(`function __fish_%[1]s_complete
  set -l args (commandline -opc) (commandline -ct)
  set -e args[1]
  "%[2]s" completion --complete fish -- $args
end
complete -a "(__fish_%[1]s_complete)" -c "%[1]s"
`, command, handler)

	case "zsh":
		scriptPath = fmt.Sprintf("zsh/site-functions/_%s", command)
		script = fmt.Sprintf(`#compdef %[1]s
_%[1]s() {
  local -a completions
  completions=(${(f)"$("%[2]s" completion --complete zsh -- "${words[@]:1:$((CURRENT-1))}")"})
  if (( ${#completions[@]} )); then
    _describe 'completions' completions
  else
    _default
  fi
}
compdef _%[1]s %[1]s
if [ "$funcstack[1]" = "_%[1]s" ]; then
  _%[1]s
fi
`, command, handler)

	case "":
		return "", "", fmt.Errorf("shell is not specified")
	default:
		return "", "", fmt.Errorf("shell %q is not supported yet, supported: bash, fish, zsh", shell)
	}

	return scriptPath, script, nil
}

func doCompletionScript(writer io.Writer, shell, command, handler string, install bool) error {
	scriptPath, script, err := generateCompletionScript(shell, command, handler)
	if err != nil {
		return err
	}
	if install {
		dataHomeDir := os.Getenv("XDG_DATA_HOME")
		if dataHomeDir == "" {
			userHomeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			dataHomeDir = filepath.Join(userHomeDir, ".local", "share")
		}
		scriptPath = filepath.Join(dataHomeDir, scriptPath)
		if _, err := fmt.Fprintf(writer, "Installing completion script: %v\n", scriptPath); err != nil {
			return err
		}
		//nolint:gosec //0o644
		if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
			return fmt.Errorf("failed to install completion script: %w", err)
		}
	} else if _, err := writer.Write([]byte(script)); err != nil {
		return err
	}
	return cli.Exit("", 0)
}

type Completor interface {
	Complete(ctx context.Context, prefix string) map[string]string
}

type Argument struct {
	Name       string
	UsageText  string
	DoComplete func(context.Context, string) map[string]string
}

var _ Completor = (*Argument)(nil)

func (a *Argument) HasName(s string) bool {
	return s == a.Name
}

func (a *Argument) Usage() string {
	return a.UsageText
}

func (a *Argument) Parse(s []string) ([]string, error) {
	return s, nil
}

func (a *Argument) Get() any {
	return nil
}

func (a *Argument) Complete(ctx context.Context, prefix string) map[string]string {
	return a.DoComplete(ctx, prefix)
}

func doCompletion(ctx context.Context, c *cli.Command, shell string, completeArgs []string) error {
	lastCmd := c.Root()
	var lastArg string

	if len(completeArgs) > 0 {
		lastArg = completeArgs[len(completeArgs)-1]

		args := append([]string{lastCmd.Name}, completeArgs...)
		if strings.HasPrefix(args[len(args)-1], "-") {
			args = args[:len(args)-1]
		}

		WithinCompletion = true
		//nolint:contextcheck //ctx
		err := lastCmd.Run(context.Background(), args)
		WithinCompletion = false

		if err == nil && LastCommand != nil {
			lastCmd = LastCommand
		} else {
			for _, arg := range completeArgs {
				if subCmd := lastCmd.Command(arg); subCmd != nil {
					lastCmd = subCmd
				} else if !strings.HasPrefix(arg, "-") {
					break
				}
			}
		}
	}

	result := map[string]string{}
	if delim := slices.Index(completeArgs, "--"); delim >= 0 && delim != len(completeArgs)-1 {
		// No completion for pass-through arguments after "--"
	} else if lastArg != "" && lastArg[0] == '-' {
		// Complete flags
		prefix := strings.TrimLeft(lastArg, "-")
		dash := lastArg[:len(lastArg)-len(prefix)]
		for _, flag := range lastCmd.Flags {
			for _, name := range flag.Names() {
				if !strings.HasPrefix(name, prefix) {
					continue
				}
				if len(name) == 1 && len(dash) > 1 {
					continue
				}
				d := dash
				if len(name) > 1 && prefix == "" && len(d) == 1 {
					d = "--"
				}
				usage := ""
				if docFlag, ok := flag.(cli.DocGenerationFlag); ok {
					usage = docFlag.GetUsage()
				}
				result[d+name] = usage
			}
		}
	} else if len(lastCmd.Commands) != 0 {
		// Complete commands
		for _, subCmd := range lastCmd.Commands {
			if subCmd.Hidden {
				continue
			}
			if strings.HasPrefix(subCmd.Name, lastArg) {
				result[subCmd.Name] = subCmd.Usage
			} else if lastArg != "" {
				for _, alias := range subCmd.Aliases {
					if strings.HasPrefix(alias, lastArg) {
						result[alias] = subCmd.Usage
					}
				}
			}
		}
	} else if args := lastCmd.Args().Slice(); len(args) > 0 && len(lastCmd.Arguments) > 0 {
		lastArg := lastCmd.Arguments[min(len(args)-1, len(lastCmd.Arguments)-1)]
		if c, ok := lastArg.(Completor); ok {
			result = c.Complete(ctx, args[len(args)-1])
		}
	}

	buffer := bufio.NewWriter(c.Writer)

	width := 0
	for suggest := range result {
		width = max(width, len(suggest))
	}

	for _, suggest := range slices.Sorted(maps.Keys(result)) {
		var err error
		usage := result[suggest]
		switch {
		case shell == "bash" && usage != "" && len(result) > 1:
			_, err = fmt.Fprintf(buffer, "%*s (%s)\n", -width-2, suggest, usage)
		case shell == "fish":
			_, err = fmt.Fprintf(buffer, "%s\t%s\n", suggest, usage)
		case shell == "zsh":
			_, err = fmt.Fprintf(buffer, "%s:%s\n", suggest, usage)
		default:
			_, err = fmt.Fprintln(buffer, suggest)
		}
		if err != nil {
			return err
		}
	}

	if err := buffer.Flush(); err != nil {
		return err
	}

	return cli.Exit("", 0)
}

func Completion() *cli.Command {
	var complete bool
	var install bool
	var command string
	var shell string
	var completeArgs []string
	return &cli.Command{
		Name:      "completion",
		Usage:     "generate shell completion",
		ArgsUsage: "[shell] [-- complete-args]...",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "complete",
				Destination: &complete,
				Usage:       "Generate completion for arguments.",
			},
			&cli.BoolFlag{
				Name:        "install",
				Destination: &install,
				Usage:       "Install shell completion script into $XDG_DATA_HOME, ~/.local/share/...",
			},
			&cli.StringFlag{
				Name:        "command",
				Destination: &command,
				Value:       "go",
				Usage:       "Command name to complete",
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "shell", Destination: &shell},
			&cli.StringArgs{Name: "complete-args", Destination: &completeArgs, Max: -1},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if WithinCompletion {
				return nil
			}
			if shell == "" {
				shell = filepath.Base(os.Getenv("SHELL"))
			}
			if !complete {
				return doCompletionScript(c.Writer, shell, command, c.Root().Name, install)
			}
			return doCompletion(ctx, c, shell, completeArgs)
		},
	}
}
