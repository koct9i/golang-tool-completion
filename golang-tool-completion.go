package main

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v3"
)

const (
	docBase = "https://pkg.go.dev/cmd/go"

	// Flag categories
	catGeneral   = "General"
	catBuild     = "Build"
	catModule    = "Modules"
	catWorkspace = "Workspaces"
	catTest      = "Testing"
	catDebug     = "Debugging"
	catOutput    = "Output"
	catTool      = "Tooling"
)

type cmdMeta struct {
	DocURL   string
	ArgsDesc string
}

func main() {

	cli.CommandHelpTemplate = `NAME:
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

	cli.SubcommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}
`

	root := &cli.Command{
		Name:      "go",
		Usage:     "Go is a tool for managing Go source code.",
		ArgsUsage: "[arguments]",
		Description: "Docs: " + docBase + "\n\n" +
			"This wrapper defines commands/flags/args for help/validation, but execution is transparent:\n" +
			"it always runs the system `go` with the original argv.\n",
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
		},
		Metadata: map[string]any{"DocURL": docBase},
		Action:   execGo,
	}

	_ = root.Run(context.Background(), os.Args)
}

// Executes the system `go` with the original argv exactly as provided.
func execGo(ctx context.Context, _ *cli.Command) error {
	args := os.Args[1:]
	c := exec.CommandContext(ctx, "go", args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return cli.Exit("", ee.ExitCode())
		}
		return err
	}
	return nil
}

func docAnchor(h string) string {
	return docBase + "#hdr-" + strings.ReplaceAll(h, " ", "_")
}

func cmdInfo(name string) cmdMeta {
	m := map[string]cmdMeta{
		"bug":       {DocURL: docAnchor("Start_a_bug_report"), ArgsDesc: "No arguments."},
		"build":     {DocURL: docAnchor("Compile_packages_and_dependencies"), ArgsDesc: "`packages` are import paths, directories, or patterns like `./...`. `.go` files from one directory are treated as one package."},
		"clean":     {DocURL: docAnchor("Remove_object_files_and_cached_files"), ArgsDesc: "`packages` are import paths/patterns like `fmt`, `./...`."},
		"doc":       {DocURL: docAnchor("Show_documentation_for_package_or_symbol"), ArgsDesc: "`query` is `package` or `[package.]symbol[.methodOrField]` (e.g. `fmt.Println`)."},
		"env":       {DocURL: docAnchor("Print_Go_environment_information"), ArgsDesc: "`var` are environment variable names (e.g. `GOPATH`). If omitted, prints many variables."},
		"fix":       {DocURL: docAnchor("Update_packages_to_use_new_APIs"), ArgsDesc: "`packages` are import paths/patterns."},
		"fmt":       {DocURL: docAnchor("Gofmt_(reformat)_package_sources"), ArgsDesc: "`packages` are import paths/patterns."},
		"generate":  {DocURL: docAnchor("Generate_Go_files_by_processing_source"), ArgsDesc: "`targets` are `.go` files or package patterns/import paths."},
		"get":       {DocURL: docAnchor("Add_dependencies_to_current_module_and_install_them"), ArgsDesc: "`packages` are package queries, optionally with `@version`."},
		"install":   {DocURL: docAnchor("Compile_and_install_packages_and_dependencies"), ArgsDesc: "`packages` are import paths/patterns (often commands)."},
		"help":      {DocURL: docBase, ArgsDesc: ""},
		"list":      {DocURL: docAnchor("List_packages_or_modules"), ArgsDesc: "`targets` are packages; with `-m`, modules."},
		"mod":       {DocURL: docAnchor("Module_maintenance"), ArgsDesc: "`command` selects a module subcommand; remaining args are subcommand args."},
		"work":      {DocURL: docAnchor("Workspace_maintenance"), ArgsDesc: "`command` selects a workspace subcommand; remaining args are subcommand args."},
		"run":       {DocURL: docAnchor("Compile_and_run_Go_program"), ArgsDesc: "`package` is import path/dir/pattern or `.go` files (single-dir). `arguments` go to the program."},
		"telemetry": {DocURL: docAnchor("Manage_telemetry_data_and_settings"), ArgsDesc: "Optional setting: `off`, `local`, or `on`."},
		"test":      {DocURL: docAnchor("Test_packages"), ArgsDesc: "`packages` are import paths/patterns; none means local directory mode. Testing flags: " + docAnchor("Testing_flags")},
		"tool":      {DocURL: docAnchor("Run_specified_go_tool"), ArgsDesc: "`command` is tool name; remaining `args` passed to it. No args lists tools."},
		"version":   {DocURL: docAnchor("Print_Go_version"), ArgsDesc: "`file` are Go binaries to inspect; none prints `go`'s own version."},
		"vet":       {DocURL: docAnchor("Report_likely_mistakes_in_packages"), ArgsDesc: "`packages` are import paths/patterns."},

		"mod download": {DocURL: docAnchor("Download_modules_to_local_cache"), ArgsDesc: "`modules` are patterns/queries like `all` or `path@version`."},
		"mod edit":     {DocURL: docAnchor("Edit_go.mod_from_tools_or_scripts"), ArgsDesc: "Optional `go.mod` path."},
		"mod graph":    {DocURL: docAnchor("Print_module_requirement_graph"), ArgsDesc: "No arguments."},
		"mod init":     {DocURL: docAnchor("Initialize_new_module_in_current_directory"), ArgsDesc: "Optional `module-path`."},
		"mod tidy":     {DocURL: docAnchor("Add_missing_and_remove_unused_modules"), ArgsDesc: "No arguments."},
		"mod vendor":   {DocURL: docAnchor("Make_vendored_copy_of_dependencies"), ArgsDesc: "No arguments."},
		"mod verify":   {DocURL: docAnchor("Verify_dependencies_have_expected_content"), ArgsDesc: "No arguments."},
		"mod why":      {DocURL: docAnchor("Explain_why_packages_or_modules_are_needed"), ArgsDesc: "`packages` are import paths/patterns (with `-m`, modules)."},

		"work edit":   {DocURL: docAnchor("Edit_go.work_from_tools_or_scripts"), ArgsDesc: "Optional `go.work` path."},
		"work init":   {DocURL: docAnchor("Initialize_workspace_file"), ArgsDesc: "`moddirs` are module directories to add as use directives."},
		"work sync":   {DocURL: docAnchor("Sync_workspace_build_list_to_modules"), ArgsDesc: "No arguments."},
		"work use":    {DocURL: docAnchor("Add_modules_to_workspace_file"), ArgsDesc: "`moddirs` are module directories to add."},
		"work vendor": {DocURL: docAnchor("Make_vendored_copy_of_dependencies"), ArgsDesc: "No arguments."},
	}
	if v, ok := m[name]; ok {
		return v
	}
	return cmdMeta{DocURL: docBase}
}

func cmdDescription(name string) string {
	meta := cmdInfo(name)
	desc := "Documentation: " + meta.DocURL
	return desc
}

func buildFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "C", Usage: "Change to dir before running the command (must be first flag).", Category: catGeneral},
		&cli.BoolFlag{Name: "a", Usage: "Force rebuilding of packages that are already up-to-date.", Category: catBuild},
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
		&cli.BoolFlag{Name: "modcacherw", Usage: "Leave newly-created module cache directories read-write.", Category: catModule},
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
		&cli.BoolFlag{Name: "modcacherw", Usage: "Leave newly-created module cache directories read-write.", Category: catModule},
		&cli.StringFlag{Name: "modfile", Usage: "Read (and possibly write) an alternate go.mod file.", Category: catModule},
	}
}

func testBinaryFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "bench", Usage: "Run only benchmarks matching regexp.", Category: catTest},
		&cli.StringFlag{Name: "benchtime", Usage: "Run enough iterations to take the specified time (e.g., 1s, 100x).", Category: catTest},
		&cli.BoolFlag{Name: "benchmem", Usage: "Print memory allocation stats for benchmarks.", Category: catTest},
		&cli.IntFlag{Name: "count", Usage: "Run each test/benchmark/fuzz seed n times.", Category: catTest},
		&cli.StringFlag{Name: "cpu", Usage: "Comma-separated list of GOMAXPROCS values.", Category: catTest},
		&cli.BoolFlag{Name: "failfast", Usage: "Do not start new tests after the first failure.", Category: catTest},
		&cli.BoolFlag{Name: "fullpath", Usage: "Show full file names in error messages.", Category: catOutput},
		&cli.StringFlag{Name: "fuzz", Usage: "Run fuzz test matching regexp.", Category: catTest},
		&cli.StringFlag{Name: "fuzztime", Usage: "Time to spend fuzzing.", Category: catTest},
		&cli.StringFlag{Name: "list", Usage: "List tests/benchmarks/examples/fuzz tests matching regexp and exit.", Category: catTest},
		&cli.IntFlag{Name: "parallel", Usage: "Maximum number of tests to run in parallel.", Category: catTest},
		&cli.StringFlag{Name: "run", Usage: "Run only tests/examples matching regexp.", Category: catTest},
		&cli.StringFlag{Name: "skip", Usage: "Skip tests matching regexp.", Category: catTest},
		&cli.BoolFlag{Name: "short", Usage: "Tell long-running tests to shorten run time.", Category: catTest},
		&cli.StringFlag{Name: "timeout", Usage: "Panic if a test runs longer than t (e.g., 10m).", Category: catTest},
		&cli.BoolFlag{Name: "v", Usage: "Verbose output: log all tests as they are run.", Category: catOutput},
		&cli.StringFlag{Name: "json", Usage: "Convert test output to JSON stream.", Category: catOutput},
		&cli.StringFlag{Name: "coverprofile", Usage: "Write a coverage profile to the named file.", Category: catTest},
		&cli.StringFlag{Name: "blockprofile", Usage: "Write a goroutine blocking profile to the named file.", Category: catDebug},
		&cli.IntFlag{Name: "blockprofilerate", Usage: "Set blocking profile rate.", Category: catDebug},
		&cli.StringFlag{Name: "cpuprofile", Usage: "Write a CPU profile to the named file.", Category: catDebug},
		&cli.StringFlag{Name: "memprofile", Usage: "Write an allocation profile to the named file.", Category: catDebug},
		&cli.IntFlag{Name: "memprofilerate", Usage: "Set memory profiling rate.", Category: catDebug},
		&cli.StringFlag{Name: "mutexprofile", Usage: "Write a mutex contention profile to the named file.", Category: catDebug},
		&cli.IntFlag{Name: "mutexprofilefraction", Usage: "Set mutex profile fraction.", Category: catDebug},
		&cli.StringFlag{Name: "trace", Usage: "Write an execution trace to the named file.", Category: catDebug},
		&cli.StringFlag{Name: "outputdir", Usage: "Write profiles to the specified directory.", Category: catOutput},
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

// https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns

// ---- commands: positional args defined DIRECTLY per command ----

func cmdBug() *cli.Command {
	return &cli.Command{
		Name:        "bug",
		Usage:       "start a bug report",
		Metadata:    map[string]any{"DocURL": docAnchor("Start_a_bug_report")},
		Description: "",
		ArgsUsage:   "",
		Arguments:   nil,
		Action:      execGo,
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
		Action:    execGo,
	}
}

func cmdClean() *cli.Command {
	return &cli.Command{
		Name:        "clean",
		Usage:       "remove object files and cached files",
		Description: cmdDescription("clean"),
		Metadata:    map[string]any{"DocURL": docAnchor("Remove_object_files_and_cached_files")},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "i", Usage: "Remove the installed packages for the named targets.", Category: catBuild},
			&cli.BoolFlag{Name: "r", Usage: "Remove obj and installed files recursively for args and deps.", Category: catBuild},
			&cli.BoolFlag{Name: "cache", Usage: "Remove all cached build and test results.", Category: catBuild},
			&cli.BoolFlag{Name: "testcache", Usage: "Expire all test results in the cache.", Category: catTest},
			&cli.BoolFlag{Name: "modcache", Usage: "Remove the entire module download cache.", Category: catModule},
			&cli.BoolFlag{Name: "fuzzcache", Usage: "Remove all cached fuzzing values.", Category: catTest},
		}, buildFlags()...),
		ArgsUsage: "[packages]",
		Arguments: []cli.Argument{argPackage()},
		Action:    execGo,
	}
}

func cmdDoc() *cli.Command {
	return &cli.Command{
		Name:        "doc",
		Usage:       "show documentation for package or symbol",
		Description: cmdDescription("doc"),
		Metadata:    map[string]any{"DocURL": docAnchor("Show_documentation_for_package_or_symbol")},
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
		Action: execGo,
	}
}

func cmdEnv() *cli.Command {
	return &cli.Command{
		Name:        "env",
		Usage:       "print Go environment information",
		Description: cmdDescription("env"),
		Metadata:    map[string]any{"DocURL": docAnchor("Print_Go_environment_information")},
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
		Action: execGo,
	}
}

func cmdFix() *cli.Command {
	return &cli.Command{
		Name:        "fix",
		Usage:       "update packages to use new APIs",
		Description: cmdDescription("fix"),
		Metadata:    map[string]any{"DocURL": docAnchor("Update_packages_to_use_new_APIs")},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "fix", Usage: "Comma-separated list of fixes to run.", Category: catGeneral},
		},
		ArgsUsage: "[packages]",
		Arguments: []cli.Argument{argPackage()},
		Action:    execGo,
	}
}

func cmdFmt() *cli.Command {
	return &cli.Command{
		Name:        "fmt",
		Usage:       "gofmt (reformat) package sources",
		Description: cmdDescription("fmt"),
		Metadata:    map[string]any{"DocURL": docAnchor("Gofmt_(reformat)_package_sources")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "n", Usage: "Print commands that would be executed.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
		},
		ArgsUsage: "[packages]",
		Arguments: []cli.Argument{argPackage()},
		Action:    execGo,
	}
}

func cmdGenerate() *cli.Command {
	return &cli.Command{
		Name:        "generate",
		Usage:       "generate Go files by processing source",
		Description: cmdDescription("generate"),
		Metadata:    map[string]any{"DocURL": docAnchor("Generate_Go_files_by_processing_source")},
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
		Action: execGo,
	}
}

func cmdGet() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Usage:       "add dependencies to current module and install them",
		Description: cmdDescription("get"),
		Metadata:    map[string]any{"DocURL": docAnchor("Add_dependencies_to_current_module_and_install_them")},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{Name: "t", Usage: "Also download test dependencies.", Category: catModule},
			&cli.BoolFlag{Name: "u", Usage: "Update modules providing dependencies.", Category: catModule},
			&cli.BoolFlag{Name: "tool", Usage: "Add packages as tool dependencies (tool directive).", Category: catModule},
		}, buildFlags()...),
		ArgsUsage: "[package@[version|latest|patch|none]]...",
		Arguments: []cli.Argument{argPackageVersion()},
		Action:    execGo,
	}
}

func cmdHelp() *cli.Command {
	return &cli.Command{
		Name:        "help",
		Usage:       "show information about command or topic",
		Description: cmdDescription("help"),
		Metadata:    map[string]any{"DocURL": docBase},
		Commands: []*cli.Command{
			{Name: "buildconstraint", Usage: "build constraints", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "buildmode", Usage: "build modes", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "c", Usage: "calling between Go and C", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "cache", Usage: "build and test caching", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "environment", Usage: "environment variables", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "filetype", Usage: "file types", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "go.mod", Usage: "the go.mod file", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "gopath", Usage: "GOPATH environment variable", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "goproxy", Usage: "module proxy protocol", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "importpath", Usage: "import path syntax", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "modules", Usage: "modules, module versions, and more", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "module-auth", Usage: "module authentication using go.sum", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "packages", Usage: "package lists and patterns", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "private", Usage: "configuration for downloading non-public code", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "testflag", Usage: "testing flags", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "testfunc", Usage: "testing functions", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
			{Name: "vcs", Usage: "controlling version control with GOVCS", Metadata: map[string]any{"DocURL": docBase}, Action: execGo},
		},
		UsageText: "go help [command|topic] [subcommand]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "query", UsageText: "help query", Min: 0, Max: -1},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			commands := []*cli.Command{}
			for _, cmd := range c.Root().Commands {
				commands = append(commands, &cli.Command{
					Name:  cmd.Name,
					Usage: cmd.Usage,
					Metadata: map[string]any{"DocURL": map[string]string{
						"bug":       docAnchor("Start_a_bug_report"),
						"build":     docAnchor("Compile_packages_and_dependencies"),
						"clean":     docAnchor("Remove_object_files_and_cached_files"),
						"doc":       docAnchor("Show_documentation_for_package_or_symbol"),
						"env":       docAnchor("Print_Go_environment_information"),
						"fix":       docAnchor("Update_packages_to_use_new_APIs"),
						"fmt":       docAnchor("Gofmt_(reformat)_package_sources"),
						"generate":  docAnchor("Generate_Go_files_by_processing_source"),
						"get":       docAnchor("Add_dependencies_to_current_module_and_install_them"),
						"help":      docBase,
						"install":   docAnchor("Compile_and_install_packages_and_dependencies"),
						"list":      docAnchor("List_packages_or_modules"),
						"mod":       docAnchor("Module_maintenance"),
						"work":      docAnchor("Workspace_maintenance"),
						"run":       docAnchor("Compile_and_run_Go_program"),
						"telemetry": docAnchor("Manage_telemetry_data_and_settings"),
						"test":      docAnchor("Test_packages"),
						"tool":      docAnchor("Run_specified_go_tool"),
						"version":   docAnchor("Print_Go_version"),
						"vet":       docAnchor("Report_likely_mistakes_in_packages"),
					}[cmd.Name]},
					Action: execGo,
				})
			}
			c.Commands = append(commands, c.Commands...)
			return ctx, nil
		},
		Action: execGo,
	}
}

func cmdInstall() *cli.Command {
	return &cli.Command{
		Name:        "install",
		Usage:       "compile and install packages and dependencies",
		Description: cmdDescription("install"),
		Metadata:    map[string]any{"DocURL": docAnchor("Compile_and_install_packages_and_dependencies")},
		Flags:       buildFlags(),
		ArgsUsage:   "[package[@version|latest]]...",
		Arguments:   []cli.Argument{argPackageVersion()},
		Action:      execGo,
	}
}

func cmdList() *cli.Command {
	return &cli.Command{
		Name:        "list",
		Usage:       "list packages or modules",
		Description: cmdDescription("list"),
		Metadata:    map[string]any{"DocURL": docAnchor("List_packages_or_modules")},
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
		Action: execGo,
	}
}

func cmdRun() *cli.Command {
	return &cli.Command{
		Name:        "run",
		Usage:       "compile and run Go program",
		Description: cmdDescription("run"),
		Metadata:    map[string]any{"DocURL": docAnchor("Compile_and_run_Go_program")},
		Flags: append([]cli.Flag{
			&cli.StringFlag{Name: "exec", Usage: "Run the generated binary under xprog (like 'time').", Category: catTool},
		}, buildFlags()...),
		ArgsUsage: "package[@version] [arguments...]",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "package", UsageText: "Program package to run"},
			&cli.StringArgs{Name: "arguments", UsageText: "Arguments passed to the compiled program", Min: 0, Max: -1},
		},
		Action: execGo,
	}
}

func cmdTelemetry() *cli.Command {
	return &cli.Command{
		Name:        "telemetry",
		Usage:       "manage telemetry data and settings",
		Description: cmdDescription("telemetry"),
		Metadata:    map[string]any{"DocURL": docAnchor("Manage_telemetry_data_and_settings")},
		ArgsUsage:   "[off|local|on]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "setting", UsageText: "Optional: off | local | on", Min: 0, Max: 1},
		},
		Action: execGo,
	}
}

func cmdTest() *cli.Command {
	return &cli.Command{
		Name:        "test",
		Usage:       "test packages",
		Description: cmdDescription("test"),
		Metadata:    map[string]any{"DocURL": docAnchor("Test_packages")},
		Flags:       append(buildFlags(), testBinaryFlags()...),
		ArgsUsage:   "[packages] [build/test flags] [test binary flags]",
		Arguments:   []cli.Argument{argPackage()},
		Action:      execGo,
	}
}

func cmdTool() *cli.Command {
	return &cli.Command{
		Name:        "tool",
		Usage:       "run specified go tool",
		Description: cmdDescription("tool"),
		Metadata:    map[string]any{"DocURL": docAnchor("Run_specified_go_tool")},
		Flags:       toolGlobalFlags(),
		Action:      execGo,
	}
}

func cmdVersion() *cli.Command {
	return &cli.Command{
		Name:        "version",
		Usage:       "print Go version",
		Description: cmdDescription("version"),
		Metadata:    map[string]any{"DocURL": docAnchor("Print_Go_version")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "m", Usage: "Print module version information (when available).", Category: catModule},
			&cli.BoolFlag{Name: "v", Usage: "Report unrecognized files when scanning directories.", Category: catOutput},
			&cli.BoolFlag{Name: "json", Usage: "Print build info as JSON (requires -m).", Category: catOutput},
		},
		ArgsUsage: "[file]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "file", UsageText: "Go binaries to inspect", Min: 0, Max: -1},
		},
		Action: execGo,
	}
}

func cmdVet() *cli.Command {
	return &cli.Command{
		Name:        "vet",
		Usage:       "report likely mistakes in packages",
		Description: cmdDescription("vet"),
		Metadata:    map[string]any{"DocURL": docAnchor("Report_likely_mistakes_in_packages")},
		Flags: append([]cli.Flag{
			&cli.StringFlag{Name: "vettool", Usage: "Use a different analysis tool.", Category: catTool},
		}, buildFlags()...),
		ArgsUsage: "[package]...",
		Arguments: []cli.Argument{argPackage()},
		Action:    execGo,
	}
}

// ---- go mod (with subcommands) ----

func cmdMod() *cli.Command {
	return &cli.Command{
		Name:        "mod",
		Usage:       "module maintenance",
		Description: cmdDescription("mod"),
		Metadata:    map[string]any{"DocURL": docAnchor("Module_maintenance")},
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
		Action: execGo,
	}
}

func cmdModDownload() *cli.Command {
	return &cli.Command{
		Name:        "download",
		Usage:       "download modules to local cache",
		Description: cmdDescription("mod download"),
		Metadata:    map[string]any{"DocURL": docAnchor("Download_modules_to_local_cache")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json", Usage: "Print JSON output.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
		},
		ArgsUsage: "package[@version]...",
		Arguments: []cli.Argument{argPackageVersion()},
		Action:    execGo,
	}
}

func cmdModEdit() *cli.Command {
	return &cli.Command{
		Name:        "edit",
		Usage:       "edit go.mod from tools or scripts",
		Description: cmdDescription("mod edit"),
		Metadata:    map[string]any{"DocURL": docAnchor("Edit_go.mod_from_tools_or_scripts")},
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
		Action: execGo,
	}
}

func cmdModGraph() *cli.Command {
	return &cli.Command{
		Name:        "graph",
		Usage:       "print module requirement graph",
		Description: cmdDescription("mod graph"),
		Metadata:    map[string]any{"DocURL": docAnchor("Print_module_requirement_graph")},
		Action:      execGo,
	}
}

func cmdModInit() *cli.Command {
	return &cli.Command{
		Name:        "init",
		Usage:       "initialize new module in current directory",
		Description: cmdDescription("mod init"),
		Metadata:    map[string]any{"DocURL": docAnchor("Initialize_new_module_in_current_directory")},
		ArgsUsage:   "[module-path]",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "module-path", UsageText: "Optional module path to initialize", Min: 0, Max: 1},
		},
		Action: execGo,
	}
}

func cmdModTidy() *cli.Command {
	return &cli.Command{
		Name:        "tidy",
		Usage:       "add missing and remove unused modules",
		Description: cmdDescription("mod tidy"),
		Metadata:    map[string]any{"DocURL": docAnchor("Add_missing_and_remove_unused_modules")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "e", Usage: "Report errors but proceed (best effort).", Category: catModule},
			&cli.BoolFlag{Name: "v", Usage: "Verbose output.", Category: catOutput},
			&cli.BoolFlag{Name: "x", Usage: "Print commands as they are executed.", Category: catOutput},
			&cli.BoolFlag{Name: "diff", Usage: "Print changes instead of applying them.", Category: catOutput},
			&cli.StringFlag{Name: "go", Usage: "Set -go=version for tidy.", Category: catModule},
			&cli.StringFlag{Name: "compat", Usage: "Set -compat=version for tidy.", Category: catModule},
		},
		Action: execGo,
	}
}

func cmdModVendor() *cli.Command {
	return &cli.Command{
		Name:        "vendor",
		Usage:       "make vendored copy of dependencies",
		Description: cmdDescription("mod vendor"),
		Metadata:    map[string]any{"DocURL": docAnchor("Make_vendored_copy_of_dependencies")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "e", Usage: "Attempt to proceed despite errors.", Category: catModule},
			&cli.BoolFlag{Name: "v", Usage: "Print names of vendored modules and packages.", Category: catOutput},
			&cli.StringFlag{Name: "o", Usage: "Output directory.", Category: catOutput},
		},
		Action: execGo,
	}
}

func cmdModVerify() *cli.Command {
	return &cli.Command{
		Name:        "verify",
		Usage:       "verify dependencies have expected content",
		Description: cmdDescription("mod verify"),
		Metadata:    map[string]any{"DocURL": docAnchor("Verify_dependencies_have_expected_content")},
		Action:      execGo,
	}
}

func cmdModWhy() *cli.Command {
	return &cli.Command{
		Name:        "why",
		Usage:       "explain why packages or modules are needed",
		Description: cmdDescription("mod why"),
		Metadata:    map[string]any{"DocURL": docAnchor("Explain_why_packages_or_modules_are_needed")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "m", Usage: "Treat arguments as modules.", Category: catModule},
		},
		ArgsUsage: "package...",
		Arguments: []cli.Argument{argPackage()},
		Action:    execGo,
	}
}

// ---- go work (with subcommands) ----

func cmdWork() *cli.Command {
	return &cli.Command{
		Name:        "work",
		Usage:       "workspace maintenance",
		Description: cmdDescription("work"),
		Metadata:    map[string]any{"DocURL": docAnchor("Workspace_maintenance")},
		Commands: []*cli.Command{
			cmdWorkEdit(),
			cmdWorkInit(),
			cmdWorkSync(),
			cmdWorkUse(),
			cmdWorkVendor(),
		},
		ArgsUsage: "<command> [argument]...",
		Action:    execGo,
	}
}

func cmdWorkEdit() *cli.Command {
	return &cli.Command{
		Name:        "edit",
		Usage:       "edit go.work from tools or scripts",
		Description: cmdDescription("work edit"),
		Metadata:    map[string]any{"DocURL": docAnchor("Edit_go.work_from_tools_or_scripts")},
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
		Action: execGo,
	}
}

func cmdWorkInit() *cli.Command {
	return &cli.Command{
		Name:        "init",
		Usage:       "initialize workspace file",
		Description: cmdDescription("work init"),
		Metadata:    map[string]any{"DocURL": docAnchor("Initialize_workspace_file")},
		ArgsUsage:   "[moddir]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "moddir", UsageText: "Module directory to add as use directives", Min: 0, Max: -1},
		},
		Action: execGo,
	}
}

func cmdWorkSync() *cli.Command {
	return &cli.Command{
		Name:        "sync",
		Usage:       "sync workspace build list to modules",
		Description: cmdDescription("work sync"),
		Metadata:    map[string]any{"DocURL": docAnchor("Sync_workspace_build_list_to_modules")},
		Action:      execGo,
	}
}

func cmdWorkUse() *cli.Command {
	return &cli.Command{
		Name:        "use",
		Usage:       "add modules to workspace file",
		Description: cmdDescription("work use"),
		Metadata:    map[string]any{"DocURL": docAnchor("Add_modules_to_workspace_file")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "r", Usage: "Search directories recursively.", Category: catWorkspace},
		},
		ArgsUsage: "[moddir]...",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "moddir", UsageText: "Module directory to add to the workspace", Min: 0, Max: -1},
		},
		Action: execGo,
	}
}

func cmdWorkVendor() *cli.Command {
	return &cli.Command{
		Name:        "vendor",
		Usage:       "make vendored copy of dependencies",
		Description: cmdDescription("work vendor"),
		Metadata:    map[string]any{"DocURL": docAnchor("Make_vendored_copy_of_dependencies")},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "e", Usage: "Attempt to proceed despite errors.", Category: catWorkspace},
			&cli.BoolFlag{Name: "v", Usage: "Print names of vendored modules and packages.", Category: catOutput},
			&cli.StringFlag{Name: "o", Usage: "Output directory.", Category: catOutput},
		},
		Action: execGo,
	}
}
