# golang-tool-completion

Shell tab-completion for the `go` command — covering subcommands, flags, package arguments, module arguments, versions, tools, and documentation symbols.

It also provides comprehensive `--help` output for command flags with links to online Go documentation.

## Overview

`golang-tool-completion` is a transparent wrapper around the standard `go` tool.

It models `go` subcommands, flags, and arguments with [urfave/cli](https://github.com/urfave/cli) to provide usage help and shell tab-completion. Normal invocations are forwarded unchanged to the real `go` binary; only `completion` and `--help` are handled by this program.

```text
go build -h          → prints complete usage and link to documentation
go build -<TAB>      → completes command flags with a short description
go get <TAB>         → completes package paths from cached and trending modules
go get <substr><TAB> → completes module paths by substring (>= 3 chars long)
go get <pkg>@<TAB>   → completes package versions
go get -u <TAB>      → completes modules already used by the current module
go get -tool <TAB>   → completes used and well-known tool packages
go doc <pkg>.<TAB>   → completes package symbols
go tool <TAB>        → completes tool command names
go install <TAB>     → completes used and well-known tool packages
```

For example `go i<TAB> gopls<TAB>@l<TAB>` turns into `go install golang.org/x/tools/gopls@latest`.

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

Argument completion for `go get` and package-oriented commands can use trending modules in addition to modules present in cache.

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

Completion for `go get -tool`, `go install`, and `go run` can suggest well-known tools in addition to main packages from cached modules and tools added into current module.

The list is read from:

```text
~/.config/golang-tool-completion/tools.txt
```

If that file does not exist, the program uses the embedded [`gomodules/tools.txt`](gomodules/tools.txt) list of tools picked from [awesome-go](https://github.com/avelino/awesome-go).

Disable tool suggestions:

```sh
golang-tool-completion completion --disable-tools
```

Add your own tool package:

```sh
golang-tool-completion completion --add-tool <package> [--description <description>]
```

## How it works

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
* **Main packages** from cached modules for `go run` and `go install`.
* **Tools** from `go tool`, `go list tool` and list of well-known tools.

## Links

- [Go command documentation](https://pkg.go.dev/cmd/go)
- [Go module cache reference](https://go.dev/ref/mod#module-cache)
- [golang proposal: bash_completions support #58598](https://github.com/golang/go/issues/58598)
- [github.com/urfave/cli](https://github.com/urfave/cli)
- [github.com/posener/complete](https://github.com/posener/complete/tree/master)
- [github.com/spf13/cobra](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md)

## License

[MIT](LICENSE)
