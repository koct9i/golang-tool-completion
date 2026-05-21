# golang-tool-completion

Shell tab-completion for the `go` command — covering every subcommand, flag, and package/module argument.

## Overview

`golang-tool-completion` is a transparent wrapper around the standard `go` tool.
It models all `go` subcommands, flags and arguments using [urfave/cli](https://github.com/urfave/cli), which gives shells
enough information to offer rich tab-completion.
Every invocation is forwarded unchanged to the real `go` binary, so the wrapper has zero impact on day-to-day usage.

```
go build -<TAB>          → lists every build flag with a short description
go get golang.org/<TAB>  → completes module paths from your local module cache
go test -run <TAB>       → lists test flags
```

## Features

| Feature | Details |
|---|---|
| **Full subcommand coverage** | `build`, `test`, `get`, `mod`, `work`, `run`, `install`, `list`, `doc`, `vet`, `fmt`, `generate`, `clean`, `fix`, `env`, `bug`, `telemetry`, `tool`, `version` |
| **Flag completion** | Every flag for every subcommand with a one-line description |
| **Package / module completion** | Completes import paths from `GOMODCACHE` (extracted *and* download trees) |
| **Popular modules** | Augmented with the top modules from [goproxy.cn/stats](https://goproxy.cn/stats/trends/last-30-days) |
| **Shell support** | bash · fish · zsh |
| **Zero overhead** | Transparent `exec` hand-off to the real `go` binary after parsing |

## Installation

```sh
go install github.com/koct9i/golang-tool-completion@latest
```

Then install the shell completion script for your current shell (`$SHELL` is detected automatically):

```sh
go run github.com/koct9i/golang-tool-completion@latest completion --install
```

Or install for a specific shell:

```sh
golang-tool-completion completion bash  --install
golang-tool-completion completion fish  --install
golang-tool-completion completion zsh   --install
```

Scripts are written to `$XDG_DATA_HOME` (defaults to `~/.local/share`):

| Shell | Path |
|---|---|
| bash | `~/.local/share/bash-completion/completions/go` |
| fish | `~/.local/share/fish/vendor_completions.d/go.fish` |
| zsh  | `~/.local/share/zsh/site-functions/_go` |

Restart your shell (or `source` the script) to activate completion.

> **Important:** place the wrapper binary first in `$PATH` so it shadows the system `go` binary.
> The wrapper detects recursion and aborts if the real `go` cannot be found.

## Manual / offline installation

Print the completion script to stdout (no `--install` flag) and save it wherever you like:

```sh
golang-tool-completion completion bash > /etc/bash_completion.d/go
golang-tool-completion completion fish > ~/.config/fish/completions/go.fish
golang-tool-completion completion zsh  > "${fpath[1]}/_go"
```

## How it works

```
shell <TAB>
  └─ golang-tool-completion completion --complete <shell> -- <words...>
       └─ parses argv, walks the command tree, emits one completion per line
shell displays completions

<Enter>
  └─ golang-tool-completion <command> [flags…]
       └─ exec(go, original argv)   ← zero overhead, same process
```

Argument completion for `build`, `get`, `install`, `list`, `run`, and similar commands is provided by reading the local module cache:

* **Extracted tree** — directories named `<escaped-module>@<version>/` under `GOMODCACHE`
* **Download tree** — `cache/download/<escaped-module>/@v/list` files

Popular modules from goproxy.cn are embedded at build time via `//go:embed` and serve as a seed when the cache is cold.

## Development

```sh
make          # fmt + vet + test + build
make install  # build, install, and register completion for the current shell
make lint     # run golangci-lint
make test     # go test ./...
```

## License

[MIT](LICENSE)
