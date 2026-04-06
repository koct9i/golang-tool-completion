package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/urfave/cli/v3"
)

const (
	docGoCmd = "https://pkg.go.dev/cmd/go"
	// docGoMod = "https://go.dev/ref/mod"
	// docGoSpec = "https://go.dev/ref/spec"

	// Flag categories
	catGeneral   = "General"
	catBuild     = "Build"
	catModule    = "Modules"
	catWorkspace = "Workspaces"
	catTest      = "Testing"
	catDebug     = "Debugging"
	catOutput    = "Output"
	catTool      = "Tooling"
	catCache     = "Cache"

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
			"DocURL": docGoCmd,
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
			cmdBug(),
			cmdBuild(),
			cmdClean(),
			cmdDoc(),
			cmdEnv(),
			cmdFix(),
			cmdFmt(),
			cmdGenerate(),
			cmdGet(),
			cmdHelp(),
			cmdInstall(),
			cmdList(),
			cmdMod(),
			cmdWork(),
			cmdRun(),
			cmdTelemetry(),
			cmdTest(),
			cmdTool(),
			cmdVersion(),
			cmdVet(),
			cmdCompletion(),
		},
		Action:   noop,
		HideHelp: true,
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

func noop(ctx context.Context, _ *cli.Command) error {
	return nil
}

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
	} else {
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

func cmdCompletion() *cli.Command {
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

func docAnchor(h string) string {
	return docGoCmd + "#hdr-" + strings.ReplaceAll(h, " ", "_")
}

func buildFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "C", Usage: "Change to dir before running the command (must be first flag).", Category: catGeneral},
		&cli.BoolFlag{Name: "a", Usage: "Force rebuilding of packages that are already up-to-date.", Category: catCache},
		&cli.BoolFlag{Name: "n", Usage: "Print the commands but do not run them.", Category: catOutput},
		&cli.IntFlag{Name: "p", Usage: "The number of programs that can be run in parallel.", Category: catBuild},
		&cli.BoolFlag{Name: "race", Usage: "Enable data race detection.", Category: catDebug},
		&cli.BoolFlag{Name: "msan", Usage: "Enable interoperation with memory sanitizer.", Category: catDebug},
		&cli.BoolFlag{Name: "asan", Usage: "Enable interoperation with address sanitizer.", Category: catDebug},
		&cli.BoolFlag{Name: "cover", Usage: "Enable code coverage instrumentation.", Category: catTest},
		&cli.StringFlag{Name: "covermode", Usage: "Coverage mode: set, count, atomic (sets -cover).", Category: catTest},
		&cli.StringFlag{Name: "coverpkg", Usage: "Comma-separated patterns of packages for which to apply coverage (sets -cover).", Category: catTest},
		&cli.BoolFlag{Name: "v", Usage: "Print the names of packages as they are compiled.", Category: catOutput},
		&cli.BoolFlag{Name: "work", Usage: "Print the name of the temporary work directory and do not delete it.", Category: catOutput},
		&cli.BoolFlag{Name: "x", Usage: "Print the commands.", Category: catOutput},
		&cli.BoolFlag{Name: "json", Usage: "Emit build output in JSON suitable for automated processing.", Category: catOutput},
		&cli.StringFlag{Name: "asmflags", Usage: "Args for each 'go tool asm' (supports [pattern=] prefix).", Category: catBuild},
		&cli.StringFlag{Name: "buildmode", Usage: "Build mode to use.", Category: catBuild},
		&cli.StringFlag{Name: "buildvcs", Usage: `Stamp binaries with VCS info: "true","false","auto".`, Category: catBuild},
		&cli.StringFlag{Name: "compiler", Usage: "Compiler to use: gc or gccgo.", Category: catBuild},
		&cli.StringFlag{Name: "gccgoflags", Usage: "Args for each gccgo compiler/linker invocation.", Category: catBuild},
		&cli.StringFlag{Name: "gcflags", Usage: "Args for each 'go tool compile' (supports [pattern=] prefix).", Category: catBuild},
		&cli.StringFlag{Name: "installsuffix", Usage: "Suffix to use in the package installation directory.", Category: catBuild},
		&cli.StringFlag{Name: "ldflags", Usage: "Args for each 'go tool link' invocation.", Category: catBuild},
		&cli.BoolFlag{Name: "linkshared", Usage: "Link against shared libraries created with -buildmode=shared.", Category: catBuild},
		&cli.StringFlag{Name: "mod", Usage: "Module download mode: readonly, vendor, or mod.", Category: catModule},
		&cli.BoolFlag{Name: "modcacherw", Usage: "Leave newly-created module cache directories read-write.", Category: catCache},
		&cli.StringFlag{Name: "modfile", Usage: "Read (and possibly write) an alternate go.mod file.", Category: catModule},
		&cli.StringFlag{Name: "overlay", Usage: "Read a JSON config file that provides an overlay for build operations.", Category: catBuild},
		&cli.StringFlag{Name: "pgo", Usage: `PGO profile file ("auto","off", or path).`, Category: catBuild},
		&cli.StringFlag{Name: "pkgdir", Usage: "Install and load packages from dir instead of the usual locations.", Category: catBuild},
		&cli.StringFlag{Name: "tags", Usage: "Comma-separated list of build tags to consider satisfied.", Category: catBuild},
		&cli.StringFlag{Name: "toolexec", Usage: "Program to invoke toolchain programs (vet/asm/compile/link).", Category: catTool},
		&cli.BoolFlag{Name: "trimpath", Usage: "Remove all file system paths from the resulting executable.", Category: catBuild},
		&cli.StringFlag{Name: "toolchain", Usage: "Select the Go toolchain to use.", Category: catBuild},
	}
}

func toolGlobalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "C", Usage: "Change to dir before running the command (must be first flag).", Category: catGeneral},
		&cli.StringFlag{Name: "overlay", Usage: "Read a JSON config file that provides an overlay for build operations.", Category: catBuild},
		&cli.BoolFlag{Name: "modcacherw", Usage: "Leave newly-created module cache directories read-write.", Category: catCache},
		&cli.StringFlag{Name: "modfile", Usage: "Read (and possibly write) an alternate go.mod file.", Category: catModule},
	}
}

func testBinaryFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "bench", Usage: "Run only benchmarks matching regexp.", Category: catTest},
		&cli.StringFlag{Name: "benchtime", Usage: "Run enough iterations to take the specified time (e.g., 1s, 100x). [cacheable]", Category: catTest},
		&cli.BoolFlag{Name: "benchmem", Usage: "Print memory allocation stats for benchmarks.", Category: catTest},
		&cli.IntFlag{Name: "count", Usage: "Run each test/benchmark/fuzz seed n times. Use -count=1 to disable caching.", Category: catTest},
		&cli.StringFlag{Name: "cpu", Usage: "Comma-separated list of GOMAXPROCS values. [cacheable]", Category: catTest},
		&cli.BoolFlag{Name: "failfast", Usage: "Do not start new tests after the first failure. [cacheable]", Category: catTest},
		&cli.BoolFlag{Name: "fullpath", Usage: "Show full file names in error messages. [cacheable]", Category: catOutput},
		&cli.StringFlag{Name: "fuzz", Usage: "Run fuzz test matching regexp.", Category: catTest},
		&cli.StringFlag{Name: "fuzztime", Usage: "Time to spend fuzzing.", Category: catTest},
		&cli.StringFlag{Name: "list", Usage: "List tests/benchmarks/examples/fuzz tests matching regexp and exit. [cacheable]", Category: catTest},
		&cli.IntFlag{Name: "parallel", Usage: "Maximum number of tests to run in parallel. [cacheable]", Category: catTest},
		&cli.StringFlag{Name: "run", Usage: "Run only tests/examples matching regexp. [cacheable]", Category: catTest},
		&cli.StringFlag{Name: "skip", Usage: "Skip tests matching regexp. [cacheable]", Category: catTest},
		&cli.BoolFlag{Name: "short", Usage: "Tell long-running tests to shorten run time. [cacheable]", Category: catTest},
		&cli.StringFlag{Name: "timeout", Usage: "Panic if a test runs longer than t (e.g., 10m). [cacheable]", Category: catTest},
		&cli.BoolFlag{Name: "v", Usage: "Verbose output: log all tests as they are run. [cacheable]", Category: catOutput},
		&cli.BoolFlag{Name: "json", Usage: "Convert test output to JSON stream. [cacheable]", Category: catOutput},
		&cli.StringFlag{Name: "coverprofile", Usage: "Write a coverage profile to the named file. [cacheable]", Category: catTest},
		&cli.StringFlag{Name: "blockprofile", Usage: "Write a goroutine blocking profile to the named file.", Category: catDebug},
		&cli.IntFlag{Name: "blockprofilerate", Usage: "Set blocking profile rate.", Category: catDebug},
		&cli.StringFlag{Name: "cpuprofile", Usage: "Write a CPU profile to the named file.", Category: catDebug},
		&cli.StringFlag{Name: "memprofile", Usage: "Write an allocation profile to the named file.", Category: catDebug},
		&cli.IntFlag{Name: "memprofilerate", Usage: "Set memory profiling rate.", Category: catDebug},
		&cli.StringFlag{Name: "mutexprofile", Usage: "Write a mutex contention profile to the named file.", Category: catDebug},
		&cli.IntFlag{Name: "mutexprofilefraction", Usage: "Set mutex profile fraction.", Category: catDebug},
		&cli.StringFlag{Name: "trace", Usage: "Write an execution trace to the named file.", Category: catDebug},
		&cli.StringFlag{Name: "outputdir", Usage: "Write profiles to the specified directory. [cacheable]", Category: catOutput},
	}
}

func argPackage() cli.Argument {
	return &cli.StringArgs{
		Name:      "package",
		UsageText: "Package, Documentation: " + docAnchor("Package_lists_and_patterns"),
		Min:       0,
		Max:       -1,
	}
}

func argPackageVersion() cli.Argument {
	return &cli.StringArgs{
		Name:      "package",
		UsageText: "Package with version, Documentation: " + docAnchor("Package_lists_and_patterns"),
		Min:       0,
		Max:       -1,
	}
}

func cmdBug() *cli.Command {
	return &cli.Command{
		Name:        "bug",
		Usage:       "start a bug report",
		Metadata:    map[string]any{"DocURL": docAnchor("Start_a_bug_report")},
		Description: "",
		ArgsUsage:   "",
		Arguments:   nil,
		Action:      noop,
	}
}

func cmdBuild() *cli.Command {
	return &cli.Command{
		Name:        "build",
		Usage:       "compile packages and dependencies",
		Metadata:    map[string]any{"DocURL": docAnchor("Compile_packages_and_dependencies")},
		Description: "",
		Flags: append([]cli.Flag{
			&cli.StringFlag{Name: "o", Usage: "Output file or directory.", Category: catOutput},
		}, buildFlags()...),
		ArgsUsage: "[packages]",
		Arguments: []cli.Argument{argPackage()},
		Action:    noop,
	}
}

func cmdClean() *cli.Command {
	return &cli.Command{
		Name:     "clean",
		Usage:    "remove object files and cached files",
		Metadata: map[string]any{"DocURL": docAnchor("Remove_object_files_and_cached_files")},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "i", Usage: "Remove the installed packages for the named targets.", Category: catCache},
			&cli.BoolFlag{Name: "r", Usage: "Remove obj and installed files recursively for args and deps.", Category: catCache},
			&cli.BoolFlag{Name: "cache", Usage: "Remove all cached build and test results.", Category: catCache},
			&cli.BoolFlag{Name: "testcache", Usage: "Expire all test results in the cache.", Category: catCache},
			&cli.BoolFlag{Name: "modcache", Usage: "Remove the entire module download cache.", Category: catCache},
			&cli.BoolFlag{Name: "fuzzcache", Usage: "Remove all cached fuzzing values.", Category: catCache},
		}, buildFlags()...),
		ArgsUsage: "[packages]",
		Arguments: []cli.Argument{argPackage()},
		Action:    noop,
	}
}

func cmdDoc() *cli.Command {
	return &cli.Command{
		Name:     "doc",
		Usage:    "show documentation for package or symbol",
		Metadata: map[string]any{"DocURL": docAnchor("Show_documentation_for_package_or_symbol")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all", Usage: "Show all the documentation for the package.", Category: catOutput},
			&cli.BoolFlag{Name: "c", Usage: "Respect case when matching symbols.", Category: catGeneral},
			&cli.BoolFlag{Name: "cmd", Usage: "Treat a command (package main) like a regular package.", Category: catGeneral},
			&cli.BoolFlag{Name: "http", Usage: "Serve HTML docs over HTTP.", Category: catTool},
			&cli.BoolFlag{Name: "short", Usage: "One-line representation for each symbol.", Category: catOutput},
			&cli.BoolFlag{Name: "src", Usage: "Show the full source code for the symbol.", Category: catOutput},
			&cli.BoolFlag{Name: "u", Usage: "Show docs for unexported symbols too.", Category: catOutput},
		},
		ArgsUsage: "package[.symbol[.methodOrField]]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "query", UsageText: "Package, symbol, method or field", Min: 0, Max: -1},
		},
		Action: noop,
	}
}

func cmdEnv() *cli.Command {
	return &cli.Command{
		Name:     "env",
		Usage:    "print Go environment information",
		Metadata: map[string]any{"DocURL": docAnchor("Print_Go_environment_information")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json", Usage: "Print environment in JSON format.", Category: catOutput},
			&cli.BoolFlag{Name: "changed", Usage: "Print only settings that differ from defaults.", Category: catOutput},
			&cli.BoolFlag{Name: "u", Usage: "Unset default settings for named variables.", Category: catGeneral},
			&cli.BoolFlag{Name: "w", Usage: "Set default settings for named variables.", Category: catGeneral},
		},
		ArgsUsage: "[NAME[=VALUE]]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "variable", UsageText: "Environment variable names (e.g. GOPATH, GOMOD)", Min: 0, Max: -1},
		},
		Action: noop,
	}
}

func cmdFix() *cli.Command {
	return &cli.Command{
		Name:     "fix",
		Usage:    "update packages to use new APIs",
		Metadata: map[string]any{"DocURL": docAnchor("Update_packages_to_use_new_APIs")},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "fix", Usage: "Comma-separated list of fixes to run.", Category: catGeneral},
		},
		ArgsUsage: "[packages]",
		Arguments: []cli.Argument{argPackage()},
		Action:    noop,
	}
}

func cmdFmt() *cli.Command {
	return &cli.Command{
		Name:     "fmt",
		Usage:    "gofmt (reformat) package sources",
		Metadata: map[string]any{"DocURL": docAnchor("Gofmt__reformat__package_sources")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "n", Usage: "Print commands that would be executed.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
		},
		ArgsUsage: "[packages]",
		Arguments: []cli.Argument{argPackage()},
		Action:    noop,
	}
}

func cmdGenerate() *cli.Command {
	return &cli.Command{
		Name:     "generate",
		Usage:    "generate Go files by processing source",
		Metadata: map[string]any{"DocURL": docAnchor("Generate_Go_files_by_processing_source")},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "run", Usage: "Run only generators matching the regexp.", Category: catGeneral},
			&cli.BoolFlag{Name: "n", Usage: "Print commands but do not run them.", Category: catOutput},
			&cli.BoolFlag{Name: "v", Usage: "Verbose output.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
			&cli.StringFlag{Name: "tags", Usage: "Comma-separated list of build tags.", Category: catBuild},
		},
		ArgsUsage: "[packages | file.go]",
		Arguments: []cli.Argument{
			argPackage(),
			&cli.StringArgs{Name: "file.go", UsageText: ""},
		},
		Action: noop,
	}
}

func cmdGet() *cli.Command {
	return &cli.Command{
		Name:     "get",
		Usage:    "add dependencies to current module and install them",
		Metadata: map[string]any{"DocURL": docAnchor("Add_dependencies_to_current_module_and_install_them")},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "t", Usage: "Also download test dependencies.", Category: catModule},
			&cli.BoolFlag{Name: "u", Usage: "Update modules providing dependencies.", Category: catModule},
			&cli.BoolFlag{Name: "tool", Usage: "Add packages as tool dependencies (tool directive).", Category: catModule},
		}, buildFlags()...),
		ArgsUsage: "[package@[version|latest|patch|none]]...",
		Arguments: []cli.Argument{argPackageVersion()},
		Action:    noop,
	}
}

func cmdHelp() *cli.Command {
	return &cli.Command{
		Name:     "help",
		Usage:    "show information about command or topic",
		Metadata: map[string]any{"DocURL": docGoCmd},
		Commands: []*cli.Command{
			{Name: "buildconstraint", Usage: "build constraints", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "buildmode", Usage: "build modes", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "c", Usage: "calling between Go and C", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "cache", Usage: "build and test caching", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "environment", Usage: "environment variables", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "filetype", Usage: "file types", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "go.mod", Usage: "the go.mod file", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "gopath", Usage: "GOPATH environment variable", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "goproxy", Usage: "module proxy protocol", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "importpath", Usage: "import path syntax", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "modules", Usage: "modules, module versions, and more", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "module-auth", Usage: "module authentication using go.sum", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "packages", Usage: "package lists and patterns", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "private", Usage: "configuration for downloading non-public code", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "testflag", Usage: "testing flags", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "testfunc", Usage: "testing functions", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
			{Name: "vcs", Usage: "controlling version control with GOVCS", Metadata: map[string]any{"DocURL": docGoCmd}, Action: noop},
		},
		UsageText: "go help [command|topic] [subcommand]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "query", UsageText: "help query", Min: 0, Max: -1},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			commands := []*cli.Command{}
			for _, cmd := range c.Root().Commands {
				commands = append(commands, &cli.Command{
					Name:   cmd.Name,
					Usage:  cmd.Usage,
					Action: noop,
				})
			}
			c.Commands = append(commands, c.Commands...)
			return ctx, nil
		},
		Action: noop,
	}
}

func cmdInstall() *cli.Command {
	return &cli.Command{
		Name:      "install",
		Usage:     "compile and install packages and dependencies",
		Metadata:  map[string]any{"DocURL": docAnchor("Compile_and_install_packages_and_dependencies")},
		Flags:     buildFlags(),
		ArgsUsage: "[package[@version|latest]]...",
		Arguments: []cli.Argument{argPackageVersion()},
		Action:    noop,
	}
}

func cmdList() *cli.Command {
	return &cli.Command{
		Name:     "list",
		Usage:    "list packages or modules",
		Metadata: map[string]any{"DocURL": docAnchor("List_packages_or_modules")},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "deps", Usage: "List dependencies of each package.", Category: catGeneral},
			&cli.StringFlag{Name: "f", Usage: "Print using a custom format.", Category: catOutput},
			&cli.BoolFlag{Name: "find", Usage: "Identify packages but do not resolve dependencies.", Category: catGeneral},
			&cli.BoolFlag{Name: "json", Usage: "Print JSON instead of text.", Category: catOutput},
			&cli.BoolFlag{Name: "m", Usage: "List modules instead of packages.", Category: catModule},
			&cli.BoolFlag{Name: "test", Usage: "Include test packages.", Category: catTest},
			&cli.BoolFlag{Name: "u", Usage: "When -m, also show available upgrades (with -versions).", Category: catModule},
			&cli.BoolFlag{Name: "retracted", Usage: "When -m, include retracted versions.", Category: catModule},
			&cli.BoolFlag{Name: "versions", Usage: "When -m, show available versions.", Category: catModule},
		}, buildFlags()...),
		ArgsUsage: "[packages]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "targets", UsageText: "Packages (or modules when -m)", Min: 0, Max: -1},
		},
		Action: noop,
	}
}

func cmdRun() *cli.Command {
	return &cli.Command{
		Name:     "run",
		Usage:    "compile and run Go program",
		Metadata: map[string]any{"DocURL": docAnchor("Compile_and_run_Go_program")},
		Flags: append([]cli.Flag{
			&cli.StringFlag{Name: "exec", Usage: "Run the generated binary under xprog (like 'time').", Category: catTool},
		}, buildFlags()...),
		ArgsUsage: "package[@version] [arguments...]",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "package", UsageText: "Program package to run"},
			&cli.StringArgs{Name: "arguments", UsageText: "Arguments passed to the compiled program", Min: 0, Max: -1},
		},
		Action: noop,
	}
}

func cmdTelemetry() *cli.Command {
	return &cli.Command{
		Name:      "telemetry",
		Usage:     "manage telemetry data and settings",
		Metadata:  map[string]any{"DocURL": docAnchor("Manage_telemetry_data_and_settings")},
		ArgsUsage: "[off|local|on]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "setting", UsageText: "Optional: off | local | on", Min: 0, Max: 1},
		},
		Action: noop,
	}
}

func cmdTest() *cli.Command {
	return &cli.Command{
		Name:      "test",
		Usage:     "test packages",
		Metadata:  map[string]any{"DocURL": docAnchor("Test_packages")},
		Flags:     append(buildFlags(), testBinaryFlags()...),
		ArgsUsage: "[packages] [build/test flags] [test binary flags]",
		Arguments: []cli.Argument{argPackage()},
		Action:    noop,
	}
}

func cmdTool() *cli.Command {
	return &cli.Command{
		Name:     "tool",
		Usage:    "run specified go tool",
		Metadata: map[string]any{"DocURL": docAnchor("Run_specified_go_tool")},
		Flags:    toolGlobalFlags(),
		Action:   noop,
	}
}

func cmdVersion() *cli.Command {
	return &cli.Command{
		Name:     "version",
		Usage:    "print Go version",
		Metadata: map[string]any{"DocURL": docAnchor("Print_Go_version")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "m", Usage: "Print module version information (when available).", Category: catModule},
			&cli.BoolFlag{Name: "v", Usage: "Report unrecognized files when scanning directories.", Category: catOutput},
			&cli.BoolFlag{Name: "json", Usage: "Print build info as JSON (requires -m).", Category: catOutput},
		},
		ArgsUsage: "[file]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "file", UsageText: "Go binaries to inspect", Min: 0, Max: -1},
		},
		Action: noop,
	}
}

func cmdVet() *cli.Command {
	return &cli.Command{
		Name:     "vet",
		Usage:    "report likely mistakes in packages",
		Metadata: map[string]any{"DocURL": docAnchor("Report_likely_mistakes_in_packages")},
		Flags: append([]cli.Flag{
			&cli.StringFlag{Name: "vettool", Usage: "Use a different analysis tool.", Category: catTool},
		}, buildFlags()...),
		ArgsUsage: "[package]...",
		Arguments: []cli.Argument{argPackage()},
		Action:    noop,
	}
}

// ---- go mod (with subcommands) ----

func cmdMod() *cli.Command {
	return &cli.Command{
		Name:     "mod",
		Usage:    "module maintenance",
		Metadata: map[string]any{"DocURL": docAnchor("Module_maintenance")},
		Commands: []*cli.Command{
			cmdModDownload(),
			cmdModEdit(),
			cmdModGraph(),
			cmdModInit(),
			cmdModTidy(),
			cmdModVendor(),
			cmdModVerify(),
			cmdModWhy(),
		},
		Action: noop,
	}
}

func cmdModDownload() *cli.Command {
	return &cli.Command{
		Name:     "download",
		Usage:    "download modules to local cache",
		Metadata: map[string]any{"DocURL": docAnchor("Download_modules_to_local_cache")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json", Usage: "Print JSON output.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
		},
		ArgsUsage: "package[@version]...",
		Arguments: []cli.Argument{argPackageVersion()},
		Action:    noop,
	}
}

func cmdModEdit() *cli.Command {
	return &cli.Command{
		Name:     "edit",
		Usage:    "edit go.mod from tools or scripts",
		Metadata: map[string]any{"DocURL": docAnchor("Edit_go.mod_from_tools_or_scripts")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "fmt", Usage: "Reformat go.mod.", Category: catModule},
			&cli.StringFlag{Name: "go", Usage: "Set the expected Go language version.", Category: catModule},
			&cli.StringFlag{Name: "toolchain", Usage: "Set the toolchain line.", Category: catModule},
			&cli.BoolFlag{Name: "print", Usage: "Print go.mod after edits.", Category: catOutput},
			&cli.BoolFlag{Name: "json", Usage: "Print go.mod after edits in JSON.", Category: catOutput},
			&cli.StringSliceFlag{Name: "require", Usage: "Add a requirement (path@version).", Category: catModule},
			&cli.StringSliceFlag{Name: "droprequire", Usage: "Drop a requirement (path).", Category: catModule},
			&cli.StringSliceFlag{Name: "replace", Usage: "Add a replace directive old[@v]=new[@v].", Category: catModule},
			&cli.StringSliceFlag{Name: "dropreplace", Usage: "Drop a replace directive old[@v].", Category: catModule},
			&cli.StringSliceFlag{Name: "exclude", Usage: "Add an exclude directive (path@version).", Category: catModule},
			&cli.StringSliceFlag{Name: "dropexclude", Usage: "Drop an exclude directive (path@version).", Category: catModule},
			&cli.StringSliceFlag{Name: "retract", Usage: "Add a retract directive (version range).", Category: catModule},
			&cli.StringSliceFlag{Name: "dropretract", Usage: "Drop a retract directive (version range).", Category: catModule},
			&cli.StringSliceFlag{Name: "tool", Usage: "Add a tool directive (path@version).", Category: catModule},
			&cli.StringSliceFlag{Name: "droptool", Usage: "Drop a tool directive (path).", Category: catModule},
		},
		ArgsUsage: "[go.mod]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "go.mod", UsageText: "Optional path to a go.mod file (default: ./go.mod)", Min: 0, Max: 1},
		},
		Action: noop,
	}
}

func cmdModGraph() *cli.Command {
	return &cli.Command{
		Name:     "graph",
		Usage:    "print module requirement graph",
		Metadata: map[string]any{"DocURL": docAnchor("Print_module_requirement_graph")},
		Action:   noop,
	}
}

func cmdModInit() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "initialize new module in current directory",
		Metadata:  map[string]any{"DocURL": docAnchor("Initialize_new_module_in_current_directory")},
		ArgsUsage: "[module-path]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "module-path", UsageText: "Optional module path to initialize", Min: 0, Max: 1},
		},
		Action: noop,
	}
}

func cmdModTidy() *cli.Command {
	return &cli.Command{
		Name:     "tidy",
		Usage:    "add missing and remove unused modules",
		Metadata: map[string]any{"DocURL": docAnchor("Add_missing_and_remove_unused_modules")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "e", Usage: "Report errors but proceed (best effort).", Category: catModule},
			&cli.BoolFlag{Name: "v", Usage: "Verbose output.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
			&cli.BoolFlag{Name: "diff", Usage: "Print changes instead of applying them.", Category: catOutput},
			&cli.StringFlag{Name: "go", Usage: "Set -go=version for tidy.", Category: catModule},
			&cli.StringFlag{Name: "compat", Usage: "Set -compat=version for tidy.", Category: catModule},
		},
		Action: noop,
	}
}

func cmdModVendor() *cli.Command {
	return &cli.Command{
		Name:     "vendor",
		Usage:    "make vendored copy of dependencies",
		Metadata: map[string]any{"DocURL": docAnchor("Make_vendored_copy_of_dependencies")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "e", Usage: "Attempt to proceed despite errors.", Category: catModule},
			&cli.BoolFlag{Name: "v", Usage: "Print names of vendored modules and packages.", Category: catOutput},
			&cli.StringFlag{Name: "o", Usage: "Output directory.", Category: catOutput},
		},
		Action: noop,
	}
}

func cmdModVerify() *cli.Command {
	return &cli.Command{
		Name:     "verify",
		Usage:    "verify dependencies have expected content",
		Metadata: map[string]any{"DocURL": docAnchor("Verify_dependencies_have_expected_content")},
		Action:   noop,
	}
}

func cmdModWhy() *cli.Command {
	return &cli.Command{
		Name:     "why",
		Usage:    "explain why packages or modules are needed",
		Metadata: map[string]any{"DocURL": docAnchor("Explain_why_packages_or_modules_are_needed")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "m", Usage: "Treat arguments as modules.", Category: catModule},
		},
		ArgsUsage: "package...",
		Arguments: []cli.Argument{argPackage()},
		Action:    noop,
	}
}

// ---- go work (with subcommands) ----

func cmdWork() *cli.Command {
	return &cli.Command{
		Name:     "work",
		Usage:    "workspace maintenance",
		Metadata: map[string]any{"DocURL": docAnchor("Workspace_maintenance")},
		Commands: []*cli.Command{
			cmdWorkEdit(),
			cmdWorkInit(),
			cmdWorkSync(),
			cmdWorkUse(),
			cmdWorkVendor(),
		},
		ArgsUsage: "<command> [argument]...",
		Action:    noop,
	}
}

func cmdWorkEdit() *cli.Command {
	return &cli.Command{
		Name:     "edit",
		Usage:    "edit go.work from tools or scripts",
		Metadata: map[string]any{"DocURL": docAnchor("Edit_go.work_from_tools_or_scripts")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "fmt", Usage: "Reformat go.work.", Category: catWorkspace},
			&cli.StringFlag{Name: "go", Usage: "Set expected Go language version.", Category: catWorkspace},
			&cli.StringFlag{Name: "toolchain", Usage: "Set toolchain name.", Category: catWorkspace},
			&cli.BoolFlag{Name: "print", Usage: "Print go.work after edits.", Category: catOutput},
			&cli.BoolFlag{Name: "json", Usage: "Print go.work after edits in JSON.", Category: catOutput},
			&cli.StringSliceFlag{Name: "use", Usage: "Add use=path directive (may repeat).", Category: catWorkspace},
			&cli.StringSliceFlag{Name: "dropuse", Usage: "Drop use=path directive (may repeat).", Category: catWorkspace},
			&cli.StringSliceFlag{Name: "replace", Usage: "Add replace old[@v]=new[@v].", Category: catWorkspace},
			&cli.StringSliceFlag{Name: "dropreplace", Usage: "Drop replace old[@v].", Category: catWorkspace},
		},
		ArgsUsage: "[go.work]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "go.work", UsageText: "Optional path to a go.work file (default: ./go.work)", Min: 0, Max: 1},
		},
		Action: noop,
	}
}

func cmdWorkInit() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "initialize workspace file",
		Metadata:  map[string]any{"DocURL": docAnchor("Initialize_workspace_file")},
		ArgsUsage: "[moddir]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "moddir", UsageText: "Module directory to add as use directives", Min: 0, Max: -1},
		},
		Action: noop,
	}
}

func cmdWorkSync() *cli.Command {
	return &cli.Command{
		Name:     "sync",
		Usage:    "sync workspace build list to modules",
		Metadata: map[string]any{"DocURL": docAnchor("Sync_workspace_build_list_to_modules")},
		Action:   noop,
	}
}

func cmdWorkUse() *cli.Command {
	return &cli.Command{
		Name:     "use",
		Usage:    "add modules to workspace file",
		Metadata: map[string]any{"DocURL": docAnchor("Add_modules_to_workspace_file")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "r", Usage: "Search directories recursively.", Category: catWorkspace},
		},
		ArgsUsage: "[moddir]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "moddir", UsageText: "Module directory to add to the workspace", Min: 0, Max: -1},
		},
		Action: noop,
	}
}

func cmdWorkVendor() *cli.Command {
	return &cli.Command{
		Name:     "vendor",
		Usage:    "make vendored copy of dependencies",
		Metadata: map[string]any{"DocURL": docAnchor("Make_vendored_copy_of_dependencies")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "e", Usage: "Attempt to proceed despite errors.", Category: catWorkspace},
			&cli.BoolFlag{Name: "v", Usage: "Print names of vendored modules and packages.", Category: catOutput},
			&cli.StringFlag{Name: "o", Usage: "Output directory.", Category: catOutput},
		},
		Action: noop,
	}
}
