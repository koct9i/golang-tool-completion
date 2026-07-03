package gomodules

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"
)

const (
	MinSubstringLen = 3
)

var (
	ChangeDirectory string
	ModFile         string
	DocPackage      string
)

func CompleteModules(ctx context.Context, result map[string]string, prefix string) {
	if strings.HasPrefix(prefix, ".") {
		return
	}
	CompleteTrending(result, prefix)
	NewModCache().CompleteModules(result, prefix)
	fixupLoneResult(result)
}

func CompletePackages(ctx context.Context, result map[string]string, prefix string) {
	if strings.HasPrefix(prefix, ".") {
		return
	}
	CompleteTrending(result, prefix)
	NewModCache().CompletePackages(result, prefix)
	fixupLoneResult(result)
}

func CompleteMainPackages(ctx context.Context, result map[string]string, prefix string) {
	if strings.HasPrefix(prefix, ".") {
		completeLocalMainPackages(ctx, result, prefix)
	} else {
		CompleteToolPackages(result, prefix)
		NewModCache().CompleteMainPackages(ctx, result, prefix)
	}
	fixupLoneResult(result)
}

func CompleteUsedModules(ctx context.Context, result map[string]string, prefix string) {
	if !strings.HasPrefix(prefix, ".") {
		NewModCache().CompleteUsedModules(ctx, result, prefix)
		fixupLoneResult(result)
	}
}

func CompleteDependencies(ctx context.Context, result map[string]string, prefix string) {
	if !strings.HasPrefix(prefix, ".") {
		completeDependencies(ctx, result, prefix)
	}
}

func CompleteDocPackage(ctx context.Context, result map[string]string, prefix string) {
	parent, tail := path.Split(prefix)
	pkgname, symbol, hasDot := strings.Cut(tail, ".")
	if hasDot && pkgname != "" {
		completePackageSymbols(ctx, result, parent+pkgname, symbol, false)
	}
	if !strings.HasPrefix(prefix, ".") && (!hasDot || parent == "") {
		completeStandardPackages(ctx, result, prefix)
		completeDependencies(ctx, result, prefix)
		if !hasDot && parent == "" {
			completePackageSymbols(ctx, result, ".", tail, true)
		}
		fixupLoneResult(result)
	}
}

func CompleteDocSymbol(ctx context.Context, result map[string]string, prefix string) {
	completePackageSymbols(ctx, result, DocPackage, prefix, true)
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

func completeDependencies(ctx context.Context, result map[string]string, prefix string) {
	if strings.HasPrefix(prefix, ".") {
		return
	}
	output, err := runGo(ctx, "list", "-deps")
	if err != nil {
		return
	}
	for line := range strings.Lines(string(output)) {
		if strings.HasPrefix(line, prefix) {
			result[strings.TrimSpace(line)] = "deps"
		}
	}
}

func completeLocalMainPackages(ctx context.Context, result map[string]string, prefix string) {
	modOutput, err := runGo(ctx, "list", "-m")
	if err != nil {
		return
	}
	modulePath := strings.TrimSpace(string(modOutput))
	output, err := runGo(ctx, "list", "-f", `{{if eq .Name "main"}}{{.ImportPath}}{{end}}`, "./...")
	if err != nil {
		return
	}
	for line := range strings.Lines(string(output)) {
		pkg := strings.TrimSpace(line)
		if pkg == "" {
			continue
		}
		suggest := "."
		if suffix, found := strings.CutPrefix(pkg, modulePath); found {
			suggest += suffix
		} else {
			suggest = pkg
		}
		if strings.HasPrefix(suggest, prefix) {
			result[suggest] = filepath.Base(suggest)
		}
	}
}

func completePackageSymbols(ctx context.Context, result map[string]string, pkg string, prefix string, symbol bool) {
	output, err := runGo(ctx, "doc", "-short", pkg)
	if err != nil {
		return
	}
	for line := range strings.Lines(string(output)) {
		if kind, tail, found := strings.Cut(line, " "); found && kind != "" && strings.HasPrefix(tail, prefix) {
			if sep := strings.IndexAny(tail, " ("); sep >= 0 {
				if symbol {
					result[tail[:sep]] = kind
				} else {
					result[pkg+"."+tail[:sep]] = kind
				}
			}
		}
	}
}

func CompleteTools(ctx context.Context, result map[string]string, prefix string) {
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

func CompleteEnv(ctx context.Context, result map[string]string, prefix string) {
	output, err := runGo(ctx, "env")
	if err != nil {
		return
	}
	for line := range strings.Lines(string(output)) {
		name, value, _ := strings.Cut(line, "=")
		if strings.HasPrefix(name, prefix) {
			result[name] = strings.Trim(value, "'\n")
		} else if strings.HasPrefix(prefix, name+"=") {
			result[prefix] = ""
			result[name+"="+strings.Trim(value, "'\n")] = "current"
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
	//nolint:gosec //command
	cmd := exec.CommandContext(ctx, "go", command)
	if ChangeDirectory != "" {
		cmd.Args = append(cmd.Args, "-C", expandPath(ChangeDirectory))
	}
	if command == "tool" && ModFile != "" {
		cmd.Args = append(cmd.Args, "-modfile", expandPath(ModFile))
	}
	cmd.Args = append(cmd.Args, args...)
	return cmd.Output()
}
