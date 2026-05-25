# golang-tool-completion

Shell tab-completion for the `go` command — covering every subcommand, flag, and package/module argument.

## Overview

`golang-tool-completion` is a transparent wrapper around the standard `go` tool.

It models all `go` subcommands, flags and arguments using [urfave/cli](https://github.com/urfave/cli) to provide help and shell tab-completion.
All invocation besides `completion` and `--help` are forwarded unchanged to the standard `go` binary.

```
go build -h          → prints complete usage and link to documentation
go build -<TAB>      → completes command flags with a short description
go get <TAB>         → completes package paths from cached and trending modules
go get <pkg>@<TAB>   → completes package version
go doc <pkg>.<TAB>   → completes package symbol
go tool <TAB>        → completes tool command name
```

## Installation

```sh
go install github.com/koct9i/golang-tool-completion@latest
```

For better `go <command> -help` add shell alias:

```sh
alias go=golang-tool-completion
```

Install shell completion script for your current shell:

```sh
golang-tool-completion completion --install
```

Or install for a specific shell:

```sh
golang-tool-completion completion --install bash
golang-tool-completion completion --install fish
golang-tool-completion completion --install zsh
```

Scripts are written to `$XDG_DATA_HOME` (defaults to `~/.local/share`):

| Shell | Path |
|---|---|
| bash | `~/.local/share/bash-completion/completions/go` |
| fish | `~/.local/share/fish/vendor_completions.d/go.fish` |
| zsh  | `~/.local/share/zsh/site-functions/_go` |

Restart your shell (or `source` the script) to activate completion.

Without `--install` it prints completion script to stdout:

```sh
golang-tool-completion completion [SHELL]
```

## How it works

```
shell <TAB>
  └─ golang-tool-completion completion --complete <shell> -- <words...>
       └─ emits completion suggestions

<Enter>
  └─ golang-tool-completion <command> [flags...]
       └─ exec(go, original argv)
```

## Packages

Argument completion for `get`, `install`, `run`, and similar commands fetch package names from:

* **Cached modules** — directories named `<module>@<version>/` under `GOMODCACHE`
* **Trending** — frequently used open-source modules

List of trending modules is read from `~/.config/golang-tool-completion/trending.txt`,
fallback is [embedded](gomodules/trending.txt) list of 1000 trending go modules from [goproxy.cn](https://goproxy.cn/stats).

To to disable completion using trending modules: `golang-tool-completion completion --disable-trending`.

## Examples

`go test -h`

```
usage: go test [build/test flags] [packages] [build/test flags & test binary flags]
Run 'go help test' and 'go help testflag' for details.
```

`golang-tool-completion test -h`

```
NAME:
   golang-tool-completion test - test packages

USAGE:
   golang-tool-completion test [packages] [build/test flags] [test binary flags]

OPTIONS:
   Build

   --asmflags string       Args for each 'go tool asm' (supports [pattern=] prefix).
   --buildmode string      Build mode to use.
   --buildvcs string       Stamp binaries with VCS info: "true","false","auto".
   --compiler string       Compiler to use: gc or gccgo.
   --gccgoflags string     Args for each gccgo compiler/linker invocation.
   --gcflags string        Args for each 'go tool compile' (supports [pattern=] prefix).
   --installsuffix string  Suffix to use in the package installation directory.
   --ldflags string        Args for each 'go tool link' invocation.
   --linkshared            Link against shared libraries created with -buildmode=shared.
   --overlay string        Read a JSON config file that provides an overlay for build operations.
   --pgo string            PGO profile file ("auto","off", or path).
   --pkgdir string         Install and load packages from dir instead of the usual locations.
   --tags string           Comma-separated list of build tags to consider satisfied.
   --toolchain string      Select the Go toolchain to use.
   --trimpath              Remove all file system paths from the resulting executable.
   -p int                  The number of programs that can be run in parallel. (default: 0)

   Cache

   --modcacherw  Leave newly-created module cache directories read-write.
   -a            Force rebuilding of packages that are already up-to-date.

   Debugging

   --asan                      Enable interoperation with address sanitizer.
   --blockprofile string       Write a goroutine blocking profile to the named file.
   --blockprofilerate int      Set blocking profile rate. (default: 0)
   --cpuprofile string         Write a CPU profile to the named file.
   --memprofile string         Write an allocation profile to the named file.
   --memprofilerate int        Set memory profiling rate. (default: 0)
   --msan                      Enable interoperation with memory sanitizer.
   --mutexprofile string       Write a mutex contention profile to the named file.
   --mutexprofilefraction int  Set mutex profile fraction. (default: 0)
   --race                      Enable data race detection.
   --trace string              Write an execution trace to the named file.

   Modules

   --mod string      Module download mode: readonly, vendor, or mod.
   --modfile string  Read (and possibly write) an alternate go.mod file.

   Output

   --fullpath          Show full file names in error messages. [cacheable]
   --json              Convert test output to JSON stream. [cacheable]
   --json              Emit build output in JSON suitable for automated processing.
   --outputdir string  Write profiles to the specified directory. [cacheable]
   --work              Print the name of the temporary work directory and do not delete it.
   -n                  Print the commands but do not run them.
   -v                  Print the names of packages as they are compiled.
   -v                  Verbose output: log all tests as they are run. [cacheable]
   -x                  Print the commands.

   Testing

   --bench string         Run only benchmarks matching regexp.
   --benchmem             Print memory allocation stats for benchmarks.
   --benchtime string     Run enough iterations to take the specified time (e.g., 1s, 100x). [cacheable]
   --count int            Run each test/benchmark/fuzz seed n times. Use -count=1 to disable caching. (default: 0)
   --cover                Enable code coverage instrumentation.
   --covermode string     Coverage mode: set, count, atomic (sets -cover).
   --coverpkg string      Comma-separated patterns of packages for which to apply coverage (sets -cover).
   --coverprofile string  Write a coverage profile to the named file. [cacheable]
   --cpu string           Comma-separated list of GOMAXPROCS values. [cacheable]
   --failfast             Do not start new tests after the first failure. [cacheable]
   --fuzz string          Run fuzz test matching regexp.
   --fuzztime string      Time to spend fuzzing.
   --list string          List tests/benchmarks/examples/fuzz tests matching regexp and exit. [cacheable]
   --parallel int         Maximum number of tests to run in parallel. [cacheable] (default: 0)
   --run string           Run only tests/examples matching regexp. [cacheable]
   --short                Tell long-running tests to shorten run time. [cacheable]
   --skip string          Skip tests matching regexp. [cacheable]
   --timeout string       Panic if a test runs longer than t (e.g., 10m). [cacheable]

   Tooling

   --toolexec string  Program to invoke toolchain programs (vet/asm/compile/link).


DOCUMENTATION:
   https://pkg.go.dev/cmd/go#hdr-Test_packages
```

## Links

- [golang proposal: bash_completions support #58598](https://github.com/golang/go/issues/58598)
- [github.com/posener/complete](https://github.com/posener/complete)

## License

[MIT](LICENSE)
