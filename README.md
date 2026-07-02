# golang-tool-completion

Shell tab-completion for the `go` command — covering every subcommand, flag, and package/module argument.

Also provides comprehensive `--help` for command flags with links to online documentation.

## Overview

`golang-tool-completion` is a transparent wrapper around the standard `go` tool.

It models all `go` subcommands, flags and arguments using [urfave/cli](https://github.com/urfave/cli) to provide help and shell tab-completion.

All invocation besides `completion` and `--help` are forwarded unchanged to the standard `go` binary.

```
go build -h          → prints complete usage and link to documentation
go build -<TAB>      → completes command flags with a short description
go get <TAB>         → completes package paths from cached and trending modules
go get <substr><TAB> → completes module paths by substring >= 3 chars long
go get <pkg>@<TAB>   → completes package version
go doc <pkg>.<TAB>   → completes package symbol
go tool <TAB>        → completes tool command name
```

For example `go i<TAB> gopls<TAB>@l<TAB>` for `go install golang.org/x/tools/gopls@latest`.

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

For example to source script: `. <(golang-tool-completion completion bash)`.

## How it works

```
shell <TAB>
  └─ golang-tool-completion completion --complete <shell> -- <words...>
       └─ emits completion suggestions

shell <Enter>
  └─ golang-tool-completion <command> [flags...]
       └─ exec(go, original argv)
```

[Shell tab-completion](completion/completion.go) is implemented from scratch.

## Packages

Argument completion for `get`, `install`, `run`, and similar commands fetches package names from:

* **Cached modules** — directories named `<module>@<version>/` under `GOMODCACHE`
* **Trending** — frequently used open-source modules for `go get`
* **Tools** — frequently used programs for `go get -tool`, `go install` and `go run`

List of trending modules is read from `~/.config/golang-tool-completion/trending.txt`,
fallback is [embedded](gomodules/trending.txt) list of 1000 trending go modules from [goproxy.cn](https://goproxy.cn/stats).

To disable trending modules in completion: `golang-tool-completion completion --disable-trending`.
To add own modules: `golang-tool-completion completion --add-trending <module> [--description <description>]`.

List of main packages for tool programs is read from `~/.config/golang-tool-completion/tools.txt`,
fallback is an [embedded](gomodules/tools.txt) list of well-known Go tools. Each entry includes a short description and, where available, a project home page.
The list is curated from project documentation plus sources such as [awesome-go](https://github.com/avelino/awesome-go), [awesome-go-tools](https://github.com/gobuild/awesome-go-tools), the [Go CodeTools wiki](https://go.dev/wiki/CodeTools), and the [`golang.org/x/tools/cmd`](https://pkg.go.dev/golang.org/x/tools/cmd) package index.

To disable tools in completion: `golang-tool-completion completion --disable-tools`.
To add own tools: `golang-tool-completion completion --add-tool <package> [--description <description>]`.

## Links

- [golang proposal: bash_completions support #58598](https://github.com/golang/go/issues/58598)
- [github.com/urfave/cli](https://github.com/urfave/cli)
- [github.com/posener/complete](https://github.com/posener/complete/tree/master)
- [github.com/spf13/cobra](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md)

## License

[MIT](LICENSE)
