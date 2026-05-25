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

Argument completion for `get`, `install`, `run`, and similar commands is provided by reading the local module cache:

* **Cached modules** — directories named `<module>@<version>/` under `GOMODCACHE`
* **Trending** — frequently used open-source modules

List of trending modules is read from `~/.config/golang-tool-completion/trending.txt`.
Fallback is embedded list of 1000 trending go modules from [goproxy.cn](https://goproxy.cn/stats).
To to disable, run: `golang-tool-completion completion --disable-trending`.

## Links

- [golang proposal: bash_completions support #58598](https://github.com/golang/go/issues/58598)
- [github.com/posener/complete](https://github.com/posener/complete)

## License

[MIT](LICENSE)
