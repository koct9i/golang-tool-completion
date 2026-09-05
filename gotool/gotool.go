package gotool

import (
	"context"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/koct9i/golang-tool-completion/completion"
	"github.com/koct9i/golang-tool-completion/gomodules"
)

const (
	DocGoCmd = "https://pkg.go.dev/cmd/go"
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
)

func docAnchor(h string) string {
	return DocGoCmd + "#hdr-" + strings.ReplaceAll(h, " ", "_")
}

func GlobalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "C",
			Usage:       "Change to dir before running the command (must be first flag).",
			Destination: &gomodules.ChangeDirectory,
		},
	}
}

func buildFlags() []cli.Flag {
	return []cli.Flag{
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
		&cli.StringFlag{
			Name:        "modfile",
			Usage:       "Read (and possibly write) an alternate go.mod file.",
			Category:    catModule,
			Destination: &gomodules.ModFile,
		},
		&cli.StringFlag{Name: "overlay", Usage: "Read a JSON config file that provides an overlay for build operations.", Category: catBuild},
		&cli.StringFlag{Name: "pgo", Usage: `PGO profile file ("auto","off", or path).`, Category: catBuild},
		&cli.StringFlag{Name: "pkgdir", Usage: "Install and load packages from dir instead of the usual locations.", Category: catBuild},
		&cli.StringFlag{Name: "tags", Usage: "Comma-separated list of build tags to consider satisfied.", Category: catBuild},
		&cli.StringFlag{Name: "toolexec", Usage: "Program to invoke toolchain programs (vet/asm/compile/link).", Category: catTool},
		&cli.BoolFlag{Name: "trimpath", Usage: "Remove all file system paths from the resulting executable.", Category: catBuild},
		&cli.StringFlag{Name: "toolchain", Usage: "Select the Go toolchain to use.", Category: catBuild},
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

func argModules() cli.Argument {
	return &completion.Argument{
		Name:       "package",
		UsageText:  "Package with version, Documentation: " + docAnchor("Package_lists_and_patterns"),
		Max:        -1,
		OnComplete: gomodules.CompleteModules,
	}
}

func argPackages() cli.Argument {
	return &completion.Argument{
		Name:       "package",
		UsageText:  "Package, Documentation: " + docAnchor("Package_lists_and_patterns"),
		Max:        -1,
		OnComplete: gomodules.CompletePackages,
	}
}

func argDependencies() cli.Argument {
	return &completion.Argument{
		Name:       "package",
		UsageText:  "Package, Documentation: " + docAnchor("Package_lists_and_patterns"),
		Max:        -1,
		OnComplete: gomodules.CompleteDependencies,
	}
}

func argSourcePackages() cli.Argument {
	return &completion.Argument{
		Name:      "package",
		UsageText: "Package, Documentation: " + docAnchor("Package_lists_and_patterns"),
		Max:       -1,
	}
}

func Bug() *cli.Command {
	return &cli.Command{
		Name:         "bug",
		Usage:        "start a bug report",
		Metadata:     map[string]any{"DocURL": docAnchor("Start_a_bug_report")},
		Description:  "",
		ArgsUsage:    "",
		Arguments:    nil,
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Build() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "compile packages and dependencies",
		Metadata: map[string]any{
			"DocURL": docAnchor("Compile_packages_and_dependencies"),
		},
		Description: "",
		Flags: append([]cli.Flag{
			&cli.StringFlag{Name: "o", Usage: "Output file or directory.", Category: catOutput},
		}, buildFlags()...),
		ArgsUsage:    "[packages]",
		Arguments:    []cli.Argument{argDependencies()},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Clean() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "remove object files and cached files",
		Metadata: map[string]any{
			"DocURL": docAnchor("Remove_object_files_and_cached_files"),
		},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "i", Usage: "Remove the installed packages for the named targets.", Category: catCache},
			&cli.BoolFlag{Name: "r", Usage: "Remove obj and installed files recursively for args and deps.", Category: catCache},
			&cli.BoolFlag{Name: "cache", Usage: "Remove all cached build and test results.", Category: catCache},
			&cli.BoolFlag{Name: "testcache", Usage: "Expire all test results in the cache.", Category: catCache},
			&cli.BoolFlag{Name: "modcache", Usage: "Remove the entire module download cache.", Category: catCache},
			&cli.BoolFlag{Name: "fuzzcache", Usage: "Remove all cached fuzzing values.", Category: catCache},
		}, buildFlags()...),
		ArgsUsage:    "[packages]",
		Arguments:    []cli.Argument{argPackages()},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Doc() *cli.Command {
	return &cli.Command{
		Name:  "doc",
		Usage: "show documentation for package or symbol",
		Metadata: map[string]any{
			"DocURL": docAnchor("Show_documentation_for_package_or_symbol"),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all", Usage: "Show all the documentation for the package.", Category: catOutput},
			&cli.BoolFlag{Name: "c", Usage: "Respect case when matching symbols.", Category: catGeneral},
			&cli.BoolFlag{Name: "cmd", Usage: "Treat a command (package main) like a regular package.", Category: catGeneral},
			&cli.BoolFlag{Name: "ex", Usage: "Include executable examples.", Category: catOutput},
			&cli.BoolFlag{Name: "http", Usage: "Serve HTML docs over HTTP.", Category: catTool},
			&cli.BoolFlag{Name: "short", Usage: "One-line representation for each symbol.", Category: catOutput},
			&cli.BoolFlag{Name: "src", Usage: "Show the full source code for the symbol.", Category: catOutput},
			&cli.BoolFlag{Name: "u", Usage: "Show docs for unexported symbols too.", Category: catOutput},
		},
		ArgsUsage: "package[.symbol[.methodOrField]] [symbol]",
		Arguments: []cli.Argument{
			&completion.Argument{
				Name:        "query",
				UsageText:   "Package, symbol, method or field",
				Max:         1,
				Destination: &gomodules.DocPackage,
				OnComplete:  gomodules.CompleteDocPackage,
			},
			&completion.Argument{
				Name:       "symbol",
				UsageText:  "Symbol",
				Max:        1,
				OnComplete: gomodules.CompleteDocSymbol,
			},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Env() *cli.Command {
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
			&completion.Argument{
				Name:       "variable",
				UsageText:  "Environment variable names (e.g. GOPATH, GOMOD)",
				Max:        -1,
				OnComplete: gomodules.CompleteEnv,
			},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Fix() *cli.Command {
	return &cli.Command{
		Name:  "fix",
		Usage: "apply fixes suggested by static checkers",
		Metadata: map[string]any{
			"DocURL": docAnchor("Apply_fixes_suggested_by_static_checkers"),
		},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "diff", Usage: "Print patch as unified diff instead of applying fixes.", Category: catOutput},
			&cli.StringFlag{Name: "fixtool", Usage: "Use a different analysis tool with alternative or additional fixers.", Category: catTool},
		}, buildFlags()...),
		ArgsUsage:    "[packages]",
		Arguments:    []cli.Argument{argSourcePackages()},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Fmt() *cli.Command {
	return &cli.Command{
		Name:  "fmt",
		Usage: "gofmt (reformat) package sources",
		Metadata: map[string]any{
			"DocURL": docAnchor("Gofmt__reformat__package_sources"),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "n", Usage: "Print commands that would be executed.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
		},
		ArgsUsage:    "[packages]",
		Arguments:    []cli.Argument{argSourcePackages()},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Generate() *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "generate Go files by processing source",
		Metadata: map[string]any{
			"DocURL": docAnchor("Generate_Go_files_by_processing_source"),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "run", Usage: "Run only generators matching the regexp.", Category: catGeneral},
			&cli.StringFlag{Name: "skip", Usage: "Skip generators matching the regexp.", Category: catGeneral},
			&cli.BoolFlag{Name: "n", Usage: "Print commands but do not run them.", Category: catOutput},
			&cli.BoolFlag{Name: "v", Usage: "Verbose output.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
			&cli.StringFlag{Name: "tags", Usage: "Comma-separated list of build tags.", Category: catBuild},
		},
		ArgsUsage:    "[packages | file.go]",
		Arguments:    []cli.Argument{argSourcePackages()},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

type getUpdateFlag struct {
	cli.StringFlag
}

// Handle "-flag" as "-flag=true".
func (f *getUpdateFlag) IsBoolFlag() bool {
	return true
}

func (f *getUpdateFlag) Complete(ctx context.Context, result map[string]string, prefix string) {
	//nolint:gocritic //prefix
	if strings.HasPrefix("-u", prefix) {
		result["-u"] = f.Usage
	}
	//nolint:gocritic //prefix
	if strings.HasPrefix("-u=patch", prefix) {
		result["-u=patch"] = "Patch-only update modules providing dependencies."
	}
}

func Get() *cli.Command {
	var tool bool
	var update string
	return &cli.Command{
		Name:  "get",
		Usage: "add dependencies to current module and install them",
		Metadata: map[string]any{
			"DocURL": docAnchor("Add_dependencies_to_current_module_and_install_them"),
		},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "t", Usage: "Also download test dependencies.", Category: catModule},
			&getUpdateFlag{StringFlag: cli.StringFlag{Name: "u", Usage: "Update modules providing dependencies.", Category: catModule, Destination: &update}},
			&cli.BoolFlag{Name: "tool", Usage: "Add packages as tool dependencies (tool directive).", Category: catModule, Destination: &tool},
		}, buildFlags()...),
		ArgsUsage: "[package@[version|latest|patch|none]]...",
		Arguments: []cli.Argument{
			&completion.Argument{
				Name:      "package",
				UsageText: "Package with version, Documentation: " + docAnchor("Package_lists_and_patterns"),
				Max:       -1,
				OnComplete: func(ctx context.Context, m map[string]string, s string) {
					if tool {
						gomodules.CompleteMainPackages(ctx, m, s)
					} else if update != "" {
						gomodules.CompleteUsedModules(ctx, m, s)
					} else {
						gomodules.CompletePackages(ctx, m, s)
					}
				},
			},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Help() *cli.Command {
	return &cli.Command{
		Name:     "help",
		Usage:    "show information about command or topic",
		Metadata: map[string]any{"DocURL": DocGoCmd},
		Commands: []*cli.Command{
			{Name: "buildconstraint", Usage: "build constraints", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "buildjson", Usage: "build -json encoding", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "buildmode", Usage: "build modes", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "c", Usage: "calling between Go and C", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "cache", Usage: "build and test caching", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "environment", Usage: "environment variables", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "filetype", Usage: "file types", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "goauth", Usage: "GOAUTH environment variable", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "go.mod", Usage: "the go.mod file", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "gopath", Usage: "GOPATH environment variable", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "goproxy", Usage: "module proxy protocol", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "importpath", Usage: "import path syntax", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "modules", Usage: "modules, module versions, and more", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "module-auth", Usage: "module authentication using go.sum", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "packages", Usage: "package lists and patterns", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "private", Usage: "configuration for downloading non-public code", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "testflag", Usage: "testing flags", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "testfunc", Usage: "testing functions", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
			{Name: "vcs", Usage: "controlling version control with GOVCS", Metadata: map[string]any{"DocURL": DocGoCmd}, Action: DummyAction},
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
					Action: DummyAction,
				})
			}
			c.Commands = append(commands, c.Commands...)
			return ctx, nil
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Install() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "compile and install packages and dependencies",
		Metadata: map[string]any{
			"DocURL": docAnchor("Compile_and_install_packages_and_dependencies"),
		},
		Flags:     buildFlags(),
		ArgsUsage: "[package[@version|latest]]...",
		Arguments: []cli.Argument{
			&completion.Argument{
				Name:       "package",
				UsageText:  "Package with version, Documentation: " + docAnchor("Package_lists_and_patterns"),
				Max:        -1,
				OnComplete: gomodules.CompleteMainPackages,
			},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func List() *cli.Command {
	var modules bool
	return &cli.Command{
		Name:  "list",
		Usage: "list packages or modules",
		Metadata: map[string]any{
			"DocURL": docAnchor("List_packages_or_modules"),
		},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "compiled", Usage: "Set CompiledGoFiles in output.", Category: catGeneral},
			&cli.BoolFlag{Name: "deps", Usage: "List dependencies of each package.", Category: catGeneral},
			&cli.BoolFlag{Name: "e", Usage: "Include erroneous packages in output.", Category: catGeneral},
			&cli.BoolFlag{Name: "export", Usage: "Set Export and BuildID fields in output.", Category: catGeneral},
			&cli.StringFlag{Name: "f", Usage: "Print using a custom format.", Category: catOutput},
			&cli.BoolFlag{Name: "find", Usage: "Identify packages but do not resolve dependencies.", Category: catGeneral},
			&cli.BoolFlag{Name: "json", Usage: "Print JSON instead of text.", Category: catOutput},
			&cli.BoolFlag{Name: "m", Usage: "List modules instead of packages.", Category: catModule, Destination: &modules},
			&cli.StringFlag{Name: "reuse", Usage: "Reuse prior -m -json output from file.", Category: catModule},
			&cli.BoolFlag{Name: "test", Usage: "Include test packages.", Category: catTest},
			&cli.BoolFlag{Name: "u", Usage: "When -m, also show available upgrades (with -versions).", Category: catModule},
			&cli.BoolFlag{Name: "retracted", Usage: "When -m, include retracted versions.", Category: catModule},
			&cli.BoolFlag{Name: "versions", Usage: "When -m, show available versions.", Category: catModule},
		}, buildFlags()...),
		ArgsUsage: "[packages]",
		Arguments: []cli.Argument{
			&completion.Argument{
				Name:      "targets",
				UsageText: "Packages (or modules when -m)",
				Max:       -1,
				OnComplete: func(ctx context.Context, result map[string]string, prefix string) {
					var patterns []string
					if modules {
						gomodules.CompleteUsedModules(ctx, result, prefix)
						patterns = []string{"all"}
					} else {
						gomodules.CompleteDependencies(ctx, result, prefix)
						patterns = []string{"all", "tool", "work", "std", "cmd"}
					}
					for _, pattern := range patterns {
						if strings.HasPrefix(pattern, prefix) {
							result[strings.TrimSpace(pattern)] = "pattern"
						}
					}
				},
			},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Run() *cli.Command {
	one := 1
	return &cli.Command{
		Name:  "run",
		Usage: "compile and run Go program",
		Metadata: map[string]any{
			"DocURL": docAnchor("Compile_and_run_Go_program"),
		},
		Flags: append([]cli.Flag{
			&cli.StringFlag{Name: "exec", Usage: "Run the generated binary under xprog (like 'time').", Category: catTool},
		}, buildFlags()...),
		ArgsUsage: "package[@version] [arguments...]",
		Arguments: []cli.Argument{
			&completion.Argument{
				Name:       "package",
				UsageText:  "Program package to run",
				Max:        1,
				OnComplete: gomodules.CompleteMainPackages,
			},
			&completion.Argument{
				Name:      "arguments",
				UsageText: "Arguments passed to the compiled program",
				Max:       -1,
			},
		},
		StopOnNthArg: &one,
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Telemetry() *cli.Command {
	return &cli.Command{
		Name:      "telemetry",
		Usage:     "manage telemetry data and settings",
		Metadata:  map[string]any{"DocURL": docAnchor("Manage_telemetry_data_and_settings")},
		ArgsUsage: "[off|local|on]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "setting", UsageText: "Optional: off | local | on", Min: 0, Max: 1},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Test() *cli.Command {
	return &cli.Command{
		Name:  "test",
		Usage: "test packages",
		Metadata: map[string]any{
			"DocURL": docAnchor("Test_packages"),
		},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "args", Usage: "Pass remaining args to the test binary.", Category: catTest},
			&cli.BoolFlag{Name: "c", Usage: "Compile test binary but do not run it.", Category: catTest},
			&cli.StringFlag{Name: "exec", Usage: "Run test binary using xprog.", Category: catTool},
			&cli.StringFlag{Name: "o", Usage: "Write test binary to file or directory.", Category: catOutput},
		}, append(buildFlags(), testBinaryFlags()...)...),
		ArgsUsage:    "[packages] [build/test flags] [test binary flags]",
		Arguments:    []cli.Argument{argSourcePackages()},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Tool() *cli.Command {
	one := 1
	return &cli.Command{
		Name:      "tool",
		Usage:     "run specified go tool",
		ArgsUsage: "command [arguments]...",
		Description: "Go ships with a number of builtin tools, and additional tools may be defined in the go.mod of the current module.\n" +
			"With no arguments it prints the list of known tools.\n" +
			"\n" +
			"For more about each builtin tool command, see 'go doc cmd/<command>'.",
		Metadata: map[string]any{
			"DocURL": docAnchor("Run_specified_go_tool"),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "n", Usage: "Print command that would be executed.", Category: catOutput},
			&cli.StringFlag{Name: "overlay", Usage: "Read a JSON config file that provides an overlay for build operations.", Category: catBuild},
			&cli.BoolFlag{Name: "modcacherw", Usage: "Leave newly-created module cache directories read-write.", Category: catCache},
			&cli.StringFlag{
				Name:        "modfile",
				Usage:       "Read (and possibly write) an alternate go.mod file.",
				Category:    catModule,
				Destination: &gomodules.ModFile,
			},
		},
		Arguments: []cli.Argument{
			&completion.Argument{
				Name:       "command",
				UsageText:  "",
				Max:        1,
				OnComplete: gomodules.CompleteTools,
			},
			&completion.Argument{
				Name:      "arguments",
				UsageText: "",
				Max:       -1,
			},
		},
		StopOnNthArg: &one,
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Version() *cli.Command {
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
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func Vet() *cli.Command {
	return &cli.Command{
		Name:  "vet",
		Usage: "report likely mistakes in packages",
		Metadata: map[string]any{
			"DocURL": docAnchor("Report_likely_mistakes_in_packages"),
		},
		Flags: append([]cli.Flag{
			&cli.IntFlag{Name: "c", Usage: "Display offending line with this many lines of context (default -1).", Category: catOutput},
			&cli.BoolFlag{Name: "diff", Usage: "Print patch as unified diff instead of applying fixes.", Category: catOutput},
			&cli.BoolFlag{Name: "fix", Usage: "Apply first suggested fix instead of printing diagnostic.", Category: catGeneral},
			&cli.BoolFlag{Name: "json", Usage: "Emit JSON output.", Category: catOutput},
			&cli.StringFlag{Name: "vettool", Usage: "Use a different analysis tool.", Category: catTool},
		}, buildFlags()...),
		ArgsUsage:    "[package]...",
		Arguments:    []cli.Argument{argSourcePackages()},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

// ---- go mod (with subcommands) ----

func Mod() *cli.Command {
	return &cli.Command{
		Name:     "mod",
		Usage:    "module maintenance",
		Metadata: map[string]any{"DocURL": docAnchor("Module_maintenance")},
		Commands: []*cli.Command{
			ModDownload(),
			ModEdit(),
			ModGraph(),
			ModInit(),
			ModTidy(),
			ModVendor(),
			ModVerify(),
			ModWhy(),
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func ModDownload() *cli.Command {
	return &cli.Command{
		Name:  "download",
		Usage: "download modules to local cache",
		Metadata: map[string]any{
			"DocURL": docAnchor("Download_modules_to_local_cache"),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json", Usage: "Print JSON output.", Category: catOutput},
			&cli.StringFlag{Name: "reuse", Usage: "Reuse previous -json output from file.", Category: catModule},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
		},
		ArgsUsage:    "package[@version]...",
		Arguments:    []cli.Argument{argModules()},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func ModEdit() *cli.Command {
	return &cli.Command{
		Name:     "edit",
		Usage:    "edit go.mod from tools or scripts",
		Metadata: map[string]any{"DocURL": docAnchor("Edit_go.mod_from_tools_or_scripts")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "fmt", Usage: "Reformat go.mod.", Category: catModule},
			&cli.StringFlag{Name: "module", Usage: "Set module path.", Category: catModule},
			&cli.StringFlag{Name: "go", Usage: "Set the expected Go language version.", Category: catModule},
			&cli.StringFlag{Name: "toolchain", Usage: "Set the toolchain line.", Category: catModule},
			&cli.StringFlag{Name: "godebug", Usage: "Add godebug key=value line.", Category: catModule},
			&cli.StringFlag{Name: "dropgodebug", Usage: "Drop godebug key.", Category: catModule},
			&cli.BoolFlag{Name: "print", Usage: "Print go.mod after edits.", Category: catOutput},
			&cli.BoolFlag{Name: "json", Usage: "Print go.mod after edits in JSON.", Category: catOutput},
			&cli.BoolFlag{Name: "n", Usage: "Print commands that would be executed.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
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
			&cli.StringSliceFlag{Name: "ignore", Usage: "Add an ignore directive (path).", Category: catModule},
			&cli.StringSliceFlag{Name: "dropignore", Usage: "Drop an ignore directive (path).", Category: catModule},
		},
		ArgsUsage: "[go.mod]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "go.mod", UsageText: "Optional path to a go.mod file (default: ./go.mod)", Min: 0, Max: 1},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func ModGraph() *cli.Command {
	return &cli.Command{
		Name:     "graph",
		Usage:    "print module requirement graph",
		Metadata: map[string]any{"DocURL": docAnchor("Print_module_requirement_graph")},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "go", Usage: "Report graph as loaded by Go version.", Category: catModule},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func ModInit() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "initialize new module in current directory",
		Metadata:  map[string]any{"DocURL": docAnchor("Initialize_new_module_in_current_directory")},
		ArgsUsage: "[module-path]",
		Arguments: []cli.Argument{
			&completion.Argument{
				Name:       "module-path",
				UsageText:  "Optional module path to initialize",
				Max:        1,
				OnComplete: gomodules.CompleteGitModule,
			},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func ModTidy() *cli.Command {
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
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func ModVendor() *cli.Command {
	return &cli.Command{
		Name:     "vendor",
		Usage:    "make vendored copy of dependencies",
		Metadata: map[string]any{"DocURL": docAnchor("Make_vendored_copy_of_dependencies")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "e", Usage: "Attempt to proceed despite errors.", Category: catModule},
			&cli.BoolFlag{Name: "v", Usage: "Print names of vendored modules and packages.", Category: catOutput},
			&cli.StringFlag{Name: "o", Usage: "Output directory.", Category: catOutput},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func ModVerify() *cli.Command {
	return &cli.Command{
		Name:         "verify",
		Usage:        "verify dependencies have expected content",
		Metadata:     map[string]any{"DocURL": docAnchor("Verify_dependencies_have_expected_content")},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func ModWhy() *cli.Command {
	return &cli.Command{
		Name:  "why",
		Usage: "explain why packages or modules are needed",
		Metadata: map[string]any{
			"DocURL": docAnchor("Explain_why_packages_or_modules_are_needed"),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "m", Usage: "Treat arguments as modules.", Category: catModule},
			&cli.BoolFlag{Name: "vendor", Usage: "Exclude tests of dependencies.", Category: catModule},
		},
		ArgsUsage:    "package...",
		Arguments:    []cli.Argument{argDependencies()},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

// ---- go work (with subcommands) ----

func Work() *cli.Command {
	return &cli.Command{
		Name:     "work",
		Usage:    "workspace maintenance",
		Metadata: map[string]any{"DocURL": docAnchor("Workspace_maintenance")},
		Commands: []*cli.Command{
			WorkEdit(),
			WorkInit(),
			WorkSync(),
			WorkUse(),
			WorkVendor(),
		},
		ArgsUsage:    "<command> [argument]...",
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func WorkEdit() *cli.Command {
	return &cli.Command{
		Name:     "edit",
		Usage:    "edit go.work from tools or scripts",
		Metadata: map[string]any{"DocURL": docAnchor("Edit_go.work_from_tools_or_scripts")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "fmt", Usage: "Reformat go.work.", Category: catWorkspace},
			&cli.StringFlag{Name: "go", Usage: "Set expected Go language version.", Category: catWorkspace},
			&cli.StringFlag{Name: "toolchain", Usage: "Set toolchain name.", Category: catWorkspace},
			&cli.StringFlag{Name: "godebug", Usage: "Add godebug key=value line.", Category: catWorkspace},
			&cli.StringFlag{Name: "dropgodebug", Usage: "Drop godebug key.", Category: catWorkspace},
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
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func WorkInit() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "initialize workspace file",
		Metadata:  map[string]any{"DocURL": docAnchor("Initialize_workspace_file")},
		ArgsUsage: "[moddir]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "moddir", UsageText: "Module directory to add as use directives", Min: 0, Max: -1},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func WorkSync() *cli.Command {
	return &cli.Command{
		Name:         "sync",
		Usage:        "sync workspace build list to modules",
		Metadata:     map[string]any{"DocURL": docAnchor("Sync_workspace_build_list_to_modules")},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func WorkUse() *cli.Command {
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
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func WorkVendor() *cli.Command {
	return &cli.Command{
		Name:     "vendor",
		Usage:    "make vendored copy of dependencies",
		Metadata: map[string]any{"DocURL": docAnchor("Make_vendored_copy_of_dependencies")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "e", Usage: "Attempt to proceed despite errors.", Category: catWorkspace},
			&cli.BoolFlag{Name: "v", Usage: "Print names of vendored modules and packages.", Category: catOutput},
			&cli.StringFlag{Name: "o", Usage: "Output directory.", Category: catOutput},
		},
		Action:       DummyAction,
		OnUsageError: NoUsageErrror,
	}
}

func DummyAction(ctx context.Context, c *cli.Command) error {
	if completion.WithinCompletion {
		completion.LastCommand = c
	}
	return nil
}

func NoUsageErrror(ctx context.Context, cmd *cli.Command, err error, isSubcommand bool) error {
	return err
}
