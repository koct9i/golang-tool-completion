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

	"github.com/koct9i/golang-tool-completion/gomodules"
)

var (
	WithinCompletion   bool
	LastCommand        *cli.Command
	ParsedArguments    []string
	ArgumentCompletors []Completor
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
	return nil
}

type Completor interface {
	Complete(ctx context.Context, prefix string) map[string]string
}

// FlagCompletableValues is implemented by flags that support optional value
// completion via the -flag=value syntax (e.g. -u=patch for go get).
type FlagCompletableValues interface {
	CompleteValues(ctx context.Context, prefix string) map[string]string
}

type Argument struct {
	Name        string
	UsageText   string
	Max         int
	Destination *string
	OnComplete  func(context.Context, string) map[string]string
}

var _ Completor = (*Argument)(nil)

func (a *Argument) HasName(s string) bool {
	return s == a.Name
}

func (a *Argument) Usage() string {
	return a.UsageText
}

func (a *Argument) Get() any {
	return nil
}

func (a *Argument) Parse(s []string) ([]string, error) {
	if a.Max == -1 {
		ParsedArguments = append(ParsedArguments, s...)
		for range s {
			ArgumentCompletors = append(ArgumentCompletors, a)
		}
		return nil, nil
	}
	if len(s) > 0 {
		if a.Destination != nil {
			*a.Destination = s[0]
		}
		ParsedArguments = append(ParsedArguments, s[0])
		ArgumentCompletors = append(ArgumentCompletors, a)
		return s[1:], nil
	}
	return s, nil
}

func (a *Argument) Complete(ctx context.Context, prefix string) map[string]string {
	if a.OnComplete != nil {
		return a.OnComplete(ctx, prefix)
	}
	return nil
}

func doCompletion(ctx context.Context, c *cli.Command, shell string, completeArgs []string) error {
	lastCmd := c.Root()
	var lastArg string

	if len(completeArgs) > 0 {
		lastArg = completeArgs[len(completeArgs)-1]

		args := append([]string{lastCmd.Name}, completeArgs...)
		if strings.HasPrefix(args[len(args)-1], "-") {
			args[len(args)-1] = "--"
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
	} else if lastArg != "" && lastArg[0] == '-' && (lastCmd.StopOnNthArg == nil || len(ParsedArguments) < *lastCmd.StopOnNthArg) {
		// Complete flags
		prefix := strings.TrimLeft(lastArg, "-")
		dash := lastArg[:len(lastArg)-len(prefix)]
		if eqIdx := strings.IndexByte(prefix, '='); eqIdx >= 0 {
			// Complete optional values for "-flagname=value" syntax
			flagName := prefix[:eqIdx]
			valuePrefix := prefix[eqIdx+1:]
			for _, flag := range lastCmd.Flags {
				for _, name := range flag.Names() {
					if name != flagName {
						continue
					}
					if cv, ok := flag.(FlagCompletableValues); ok {
						for value, usage := range cv.CompleteValues(ctx, valuePrefix) {
							result[dash+flagName+"="+value] = usage
						}
					}
				}
			}
		} else {
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
	} else if len(ArgumentCompletors) > 0 && ParsedArguments[len(ParsedArguments)-1] == lastArg {
		// Complete arguments
		result = ArgumentCompletors[len(ArgumentCompletors)-1].Complete(ctx, ParsedArguments[len(ParsedArguments)-1])
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

	return nil
}

func Completion() *cli.Command {
	var complete bool
	var install bool
	var command string
	var shell string
	var completeArgs []string
	var disableTrendingModules bool
	var addTrendingModules []string
	var descriptions []string
	return &cli.Command{
		Name:      "completion",
		Usage:     "generate shell completion",
		ArgsUsage: "[options]... [shell] [-- complete-args]...",
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
			&cli.BoolFlag{
				Name:        "disable-trending",
				Usage:       "Clear list of trending modules used for completion",
				Category:    "Trending",
				Destination: &disableTrendingModules,
			},
			&cli.StringFlag{
				Name:     "add-trending",
				Usage:    "Add modules with descriptions into list of trending modules",
				Category: "Trending",
				Action: func(ctx context.Context, c *cli.Command, arg string) error {
					addTrendingModules = append(addTrendingModules, arg)
					return nil
				},
			},
			&cli.StringFlag{
				Name:     "description",
				Usage:    "Set description for added trending modules",
				Category: "Trending",
				Action: func(ctx context.Context, c *cli.Command, arg string) error {
					descriptions = append(descriptions, arg)
					return nil
				},
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "shell", Destination: &shell},
			&cli.StringArgs{Name: "complete-args", Destination: &completeArgs, Max: -1},
		},
		Action: func(ctx context.Context, c *cli.Command) (err error) {
			if WithinCompletion {
				return nil
			}
			defer func() {
				if err == nil {
					err = cli.Exit("", 0) // Do not call go tool.
				}
			}()
			if disableTrendingModules {
				return gomodules.DisableTrending()
			}
			if len(addTrendingModules) > 0 {
				return gomodules.AddTrending(addTrendingModules, descriptions)
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
