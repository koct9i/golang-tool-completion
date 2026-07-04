# golang-tool-completion

Shell tab-completion for the `go` command — covering subcommands, flags, package arguments, module arguments, versions, tools, and documentation symbols.

It also provides comprehensive `--help` output for command flags with links to online Go documentation.

## What it is

`golang-tool-completion` is a transparent wrapper around the standard `go` tool.

It models `go` subcommands, flags, and arguments with [urfave/cli](https://github.com/urfave/cli) so shells can ask it for context-aware completions. Normal invocations are forwarded unchanged to the real `go` binary; only `completion` and `--help` are handled by this program.

```text
go build -h          → prints complete usage and link to documentation
go build -<TAB>      → completes command flags with a short description
go get <TAB>         → completes package paths from cached and trending modules
go get <substr><TAB> → completes module paths by substring, once the substring is at least 3 chars
go get <pkg>@<TAB>   → completes package versions
go get -u <TAB>      → completes modules already used by the current module
go get -tool <TAB>   → completes used and well-known tool packages
go doc <pkg>.<TAB>   → completes package symbols
go tool <TAB>        → completes tool command names
go install <TAB>     → completes used and well-known tool packages
```

For example, type `go i<TAB> gopls<TAB>@l<TAB>` to expand toward:

```sh
go install golang.org/x/tools/gopls@latest
```

## User scenarios

| Scenario | Example | Completion source |
|---|---|---|
| Discover flags for a command | `go test -<TAB>` | Static model of Go command flags |
| Complete packages already in your build graph | `go test example.com/project/p<TAB>` | `go list -deps` in the current module |
| Add or upgrade a dependency | `go get golang.org/x/sync@<TAB>` | Module cache version lists plus `latest` and `patch` |
| Upgrade a dependency that is already used | `go get -u example.com/dep@v<TAB>` | `go list -m all`, including modules with `replace` directives |
| Install a tool | `go install gopls<TAB>@latest` | Configured tools list, embedded well-known tools, and cached main packages |
| Run a local tool package | `go run ./cmd/<TAB>` | `go list` over local main packages |
| Inspect documentation | `go doc fmt.P<TAB>` | `go doc -short` package symbol output |
| Explore cached packages while offline | `go list example.com/mod/p<TAB>` | Directories in `GOMODCACHE` |

## Installation

```sh
go install github.com/koct9i/golang-tool-completion@latest
```

For better `go <command> -help` output, add a shell alias:

```sh
alias go=golang-tool-completion
```

Install a shell completion script for your current shell:

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

Restart your shell, or source the generated script, to activate completion.

Without `--install`, the command prints a completion script to stdout:

```sh
golang-tool-completion completion [SHELL]
```

For example, to source bash completion without installing it:

```sh
. <(golang-tool-completion completion bash)
```

## Configuration

### Trending modules

Argument completion for `go get` and package-oriented commands can use trending modules when a module is not already present in your cache.

The list is read from:

```text
~/.config/golang-tool-completion/trending.txt
```

If that file does not exist, the program uses the embedded [`gomodules/trending.txt`](gomodules/trending.txt) list of 1000 trending Go modules from [goproxy.cn](https://goproxy.cn/stats).

Disable trending modules:

```sh
golang-tool-completion completion --disable-trending
```

Add your own module:

```sh
golang-tool-completion completion --add-trending <module> [--description <description>]
```

### Tool packages

Completion for `go get -tool`, `go install`, and `go run` can suggest main packages for tool programs.

The list is read from:

```text
~/.config/golang-tool-completion/tools.txt
```

If that file does not exist, the program uses the embedded [`gomodules/tools.txt`](gomodules/tools.txt) list of well-known tools picked from [awesome-go](https://github.com/avelino/awesome-go).

Disable tool suggestions:

```sh
golang-tool-completion completion --disable-tools
```

Add your own tool package:

```sh
golang-tool-completion completion --add-tool <package> [--description <description>]
```

## How completion works

```text
shell <TAB>
  └─ golang-tool-completion completion --complete <shell> -- <words...>
       └─ emits completion suggestions

shell <Enter>
  └─ golang-tool-completion <command> [flags...]
       └─ exec(go, original argv)
```

[Shell tab-completion](completion/completion.go) is implemented from scratch.

Package and module completion combines several sources:

* **Local module data** from `go list -deps`, `go list -m all`, `go list ./...`, and `go doc -short`.
* **Cached modules** from directories named `<module>@<version>/` under `GOMODCACHE`.
* **Cached module versions** from `GOMODCACHE/cache/download/<module>/@v/list`.
* **Trending modules** for discovery-oriented `go get` workflows.
* **Tool packages** for common installable commands.

## Notes and limitations

* The wrapper forwards normal `go` invocations unchanged, so build/test behavior should match the standard Go command.
* Completion quality improves as your module cache fills up.
* Network access is not required for cache-based completion, but the real `go` command may download modules when you execute normal Go commands.
* Substring matching starts at 3 characters to avoid very noisy suggestions.
* Directory-style suggestions end with `/`; versionable module/package suggestions end with `@`.

## Links

- [Go command documentation](https://pkg.go.dev/cmd/go)
- [Go module cache reference](https://go.dev/ref/mod#module-cache)
- [golang proposal: bash_completions support #58598](https://github.com/golang/go/issues/58598)
- [github.com/urfave/cli](https://github.com/urfave/cli)
- [github.com/posener/complete](https://github.com/posener/complete/tree/master)
- [github.com/spf13/cobra](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md)

## License

[MIT](LICENSE)
