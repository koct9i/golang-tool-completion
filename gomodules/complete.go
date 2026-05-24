package gomodules

import (
	"context"
	"os"
	"os/exec"
	"path"
	"strings"

	"golang.org/x/mod/module"
)

var (
	ChangeDirectory string
	ModFile         string
)

func CompleteModules(ctx context.Context, prefix string) map[string]string {
	if strings.HasPrefix(prefix, ".") {
		return nil
	}
	result := make(map[string]string)
	CompleteTrending(result, prefix)
	NewModCache().CompleteModules(result, prefix)
	fixupLoneResult(result)
	return result
}

func CompletePackages(ctx context.Context, prefix string) map[string]string {
	if strings.HasPrefix(prefix, ".") {
		return nil
	}
	result := make(map[string]string)
	CompleteTrending(result, prefix)
	NewModCache().CompletePackages(result, prefix)
	fixupLoneResult(result)
	return result
}

func CompleteDocPackages(ctx context.Context, prefix string) map[string]string {
	result := make(map[string]string)
	parent, tail := path.Split(prefix)
	if pkgname, symbol, ok := strings.Cut(tail, "."); ok {
		completePackageSymbols(ctx, result, parent+pkgname, symbol)
	} else if !strings.HasPrefix(prefix, ".") {
		completeStandardPackages(ctx, result, prefix)
		NewModCache().CompletePackages(result, prefix)
		fixupLoneResult(result)
	}
	return result
}

func CompleteTools(ctx context.Context, prefix string) map[string]string {
	result := make(map[string]string)
	completeTools(ctx, result, prefix)
	return result
}

func completeStandardPackages(ctx context.Context, result map[string]string, prefix string) {
	if strings.Contains(prefix, ".") {
		return
	}
	output, err := runGo(ctx, "list", "std")
	if err != nil {
		return
	}
	for line := range strings.Lines(string(output)) {
		if strings.HasPrefix(line, prefix) {
			result[strings.TrimSpace(line)] = "std"
		}
	}
}

func completePackageSymbols(ctx context.Context, result map[string]string, pkg string, prefix string) {
	output, err := runGo(ctx, "doc", "-short", pkg)
	if err != nil {
		return
	}
	for line := range strings.Lines(string(output)) {
		if kind, tail, found := strings.Cut(line, " "); found && kind != "" && strings.HasPrefix(tail, prefix) {
			if sep := strings.IndexAny(tail, " ("); sep >= 0 {
				result[pkg+"."+tail[:sep]] = kind
			}
		}
	}
}

func completeTools(ctx context.Context, result map[string]string, prefix string) {
	output, err := runGo(ctx, "tool")
	if err != nil {
		return
	}
	for line := range strings.Lines(string(output)) {
		tool := strings.TrimSpace(line)
		if strings.HasPrefix(tool, prefix) {
			//TODO: Add description for builtin tools.
			result[tool] = ""
		}
		if p, _, ok := module.SplitPathVersion(tool); ok {
			name := path.Base(p)
			if strings.HasPrefix(name, prefix) {
				if _, found := result[name]; !found {
					result[name] = tool
				}
			}
		}
	}
}

func fixupLoneResult(result map[string]string) {
	if len(result) == 1 {
		for pkg, status := range result {
			if strings.HasSuffix(pkg, "/") || strings.HasSuffix(pkg, "@") {
				result[pkg[:len(pkg)-1]] = status
			}
			break
		}
	}
}

func expandPath(arg string) string {
	if p, found := strings.CutPrefix(arg, "~/"); found {
		arg = "${HOME}/" + p
	}
	return os.ExpandEnv(arg)
}

func runGo(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "go")
	cmd.Args = append(cmd.Args, command)
	if ChangeDirectory != "" {
		cmd.Args = append(cmd.Args, "-C", expandPath(ChangeDirectory))
	}
	if command == "tool" && ModFile != "" {
		cmd.Args = append(cmd.Args, "-modfile", expandPath(ModFile))
	}
	cmd.Args = append(cmd.Args, args...)
	return cmd.Output()
}
