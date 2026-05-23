package gomodules

import (
	"os/exec"
	"path"
	"strings"

	"golang.org/x/mod/module"
)

func CompleteModules(prefix string) map[string]string {
	if strings.HasPrefix(prefix, ".") {
		return nil
	}
	result := make(map[string]string)
	CompleteTrending(result, prefix)
	NewModCache().CompleteModules(result, prefix)
	fixupLoneResult(result)
	return result
}

func CompletePackages(prefix string) map[string]string {
	if strings.HasPrefix(prefix, ".") {
		return nil
	}
	result := make(map[string]string)
	CompleteTrending(result, prefix)
	NewModCache().CompletePackages(result, prefix)
	fixupLoneResult(result)
	return result
}

func CompleteDocPackages(prefix string) map[string]string {
	result := make(map[string]string)
	parent, tail := path.Split(prefix)
	if pkgname, symbol, ok := strings.Cut(tail, "."); ok {
		completePackageSymbols(result, parent+pkgname, symbol)
	} else if !strings.HasPrefix(prefix, ".") {
		completeStandardPackages(result, prefix)
		NewModCache().CompletePackages(result, prefix)
		fixupLoneResult(result)
	}
	return result
}

func CompleteTools(prefix string) map[string]string {
	result := make(map[string]string)
	completeTools(result, prefix)
	return result
}

func completeStandardPackages(result map[string]string, prefix string) {
	if strings.Contains(prefix, ".") {
		return
	}
	cmd := exec.Command("go", "list", "std")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	for line := range strings.Lines(string(output)) {
		if strings.HasPrefix(line, prefix) {
			result[strings.TrimSpace(line)] = "std"
		}
	}
}

func completePackageSymbols(result map[string]string, pkg string, prefix string) {
	cmd := exec.Command("go", "doc", "-short", pkg)
	output, err := cmd.Output()
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

func completeTools(result map[string]string, prefix string) {
	// TODO: Forward "--modfile".
	cmd := exec.Command("go", "tool")
	output, err := cmd.Output()
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
