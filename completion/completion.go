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

func generateCompletionScript(program, shell string) (scriptPath string, script string, err error) {
	switch shell {
	case "bash":
		scriptPath = fmt.Sprintf("bash-completion/completions/%s", program)
		script = fmt.Sprintf(`__%s_complete_bash() {
  mapfile -t COMPREPLY < <("${COMP_WORDS[0]}" completion --complete bash -- "${COMP_WORDS[@]:1:COMP_CWORD}")
}
complete -o bashdefault -o default -F __%[1]s_complete_bash %[1]s
`, program)

	case "fish":
		scriptPath = fmt.Sprintf("fish/vendor_completions.d/%s.fish", program)
		script = fmt.Sprintf(`function __fish_%[1]s_complete
  set -l args (commandline -opc) (commandline -ct)
  set -e args[1]
  %[1]s completion --complete fish -- $args
end
complete -c %[1]s -a "(__fish_%[1]s_complete)"
`, program)

	case "zsh":
		scriptPath = fmt.Sprintf("zsh/site-functions/_%s", program)
		script = fmt.Sprintf(`#compdef %[1]s
_%[1]s() {
  local -a completions
  completions=(${(f)"$("${words[1]}" completion --complete zsh -- "${words[@]:1:$((CURRENT-1))}")"})
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
`, program)

	case "":
		return "", "", fmt.Errorf("Shell is not specified")
	default:
		return "", "", fmt.Errorf("Shell %q is not supported yet. Choose: bash, fish, zsh", shell)
	}

	return scriptPath, script, nil
}

func doCompletionScript(writer io.Writer, app, shell string, install bool) error {
	scriptPath, script, err := generateCompletionScript(app, shell)
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
			dataHomeDir = filepath.Join(userHomeDir, ".local/share")
		}
		scriptPath = filepath.Join(dataHomeDir, scriptPath)
		fmt.Fprintf(writer, "Installing completion script: %v\n", scriptPath)
		if err = os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
			return fmt.Errorf("Failed to install completion script: %w", err)
		}
	} else if _, err = writer.Write([]byte(script)); err != nil {
		return err
	}
	return cli.Exit("", 0)
}

func doCompletion(ctx context.Context, c *cli.Command, shell string, completeArgs []string) error {
	lastCmd := c.Root()
	for _, arg := range completeArgs {
		if subCmd := lastCmd.Command(arg); subCmd != nil {
			lastCmd = subCmd
		} else {
			break
		}
	}

	var lastArg string
	if len(completeArgs) > 0 {
		lastArg = completeArgs[len(completeArgs)-1]
	}

	result := map[string]string{}
	if delim := slices.Index(completeArgs, "--"); delim >= 0 && delim != len(completeArgs)-1 {
		// No completion for pass-through arguments after "--"
	} else if len(lastArg) > 0 && lastArg[0] == '-' {
		// Complete flags
		prefix := strings.TrimLeft(lastArg, "-")
		dash := lastArg[:len(lastArg)-len(prefix)]
		for _, flag := range lastCmd.Flags {
			for _, name := range flag.Names() {
				if strings.HasPrefix(name, prefix) {
					if len(name) == 1 && len(dash) > 1 {
						continue
					}
					d := dash
					if len(name) > 1 && len(prefix) == 0 && len(d) == 1 {
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
			} else if len(lastArg) > 0 {
				for _, alias := range subCmd.Aliases {
					if strings.HasPrefix(subCmd.Name, lastArg) {
						result[alias] = subCmd.Usage
					}
				}
			}
		}
	} else if completeArguments, ok := lastCmd.Metadata["CompleteArguments"]; ok {
		// Complete arguments
		result = completeArguments.(func(string) map[string]string)(lastArg)
	}

	buffer := bufio.NewWriter(c.Writer)
	defer buffer.Flush()

	width := 0
	for suggest := range result {
		width = max(width, len(suggest))
	}

	for _, suggest := range slices.Sorted(maps.Keys(result)) {
		usage := result[suggest]
		switch {
		case shell == "bash" && usage != "" && len(result) > 1:
			fmt.Fprintf(buffer, "%*s (%s)\n", -width-2, suggest, usage)
		case shell == "fish":
			fmt.Fprintf(buffer, "%s\t%s\n", suggest, usage)
		case shell == "zsh":
			fmt.Fprintf(buffer, "%s:%s\n", suggest, usage)
		default:
			fmt.Fprintln(buffer, suggest)
		}
	}

	return cli.Exit("", 0)
}

func Completion() *cli.Command {
	var complete bool
	var install bool
	var shell string
	var completeArgs []string
	return &cli.Command{
		Name:      "completion",
		Usage:     "generate shell completion",
		ArgsUsage: "[shell] [-- complete args]...",
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
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "shell", Destination: &shell},
			&cli.StringArgs{Name: "complete-args", Destination: &completeArgs, Max: -1},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if shell == "" {
				shell = filepath.Base(os.Getenv("SHELL"))
			}
			if !complete {
				return doCompletionScript(c.Writer, c.Root().Name, shell, install)
			}
			return doCompletion(ctx, c, shell, completeArgs)
		},
	}
}
