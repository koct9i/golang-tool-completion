package gomodules

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"golang.org/x/mod/module"
)

func CompleteModules(prefix string) map[string]string {
	result := make(map[string]string)
	CompletePopular(result, prefix)
	NewModCache().CompleteModules(result, prefix)
	fixupLoneResult(result)
	return result
}

func CompletePackages(prefix string) map[string]string {
	result := make(map[string]string)
	CompletePopular(result, prefix)
	NewModCache().CompletePackages(result, prefix)
	fixupLoneResult(result)
	return result
}

func CompleteDocPackages(prefix string) map[string]string {
	result := make(map[string]string)
	if pkg, symbolPrefix, isSymbol := splitDocSymbolPrefix(prefix); isSymbol {
		completePackageSymbols(result, pkg, symbolPrefix)
		fixupLoneResult(result)
		return result
	}
	CompletePopular(result, prefix)
	completeStandardPackages(result, prefix)
	completeLocalPackages(result, prefix)
	NewModCache().CompletePackages(result, prefix)
	fixupLoneResult(result)
	return result
}

func splitDocSymbolPrefix(prefix string) (pkg string, symbolPrefix string, ok bool) {
	lastSlash := strings.LastIndex(prefix, "/")
	lastElem := prefix[lastSlash+1:]
	dot := strings.Index(lastElem, ".")
	if dot < 0 {
		return "", "", false
	}
	pkg = prefix[:lastSlash+1+dot]
	symbolPrefix = lastElem[dot+1:]
	return pkg, symbolPrefix, true
}

func completePackageSymbols(result map[string]string, pkg string, symbolPrefix string) {
	cmd := exec.Command("go", "doc", "-short", pkg)
	output, err := cmd.Output()
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for _, symbol := range extractDocSymbols(line) {
			if strings.HasPrefix(symbol, symbolPrefix) {
				result[pkg+"."+symbol] = "symbol"
			}
		}
	}
}

func extractDocSymbols(line string) []string {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}
	if fields[0] == "func" {
		if fields[1] == "(" && len(fields) >= 4 {
			recv := strings.Trim(fields[2], "()*")
			method := strings.SplitN(fields[3], "(", 2)[0]
			if recv != "" && method != "" {
				return []string{recv + "." + method, method}
			}
			return nil
		}
		name := strings.SplitN(fields[1], "(", 2)[0]
		if name != "" {
			return []string{name}
		}
		return nil
	}
	if fields[0] == "type" || fields[0] == "var" || fields[0] == "const" {
		if fields[1] != "(" {
			return []string{fields[1]}
		}
	}
	return nil
}

var (
	stdPackagesOnce sync.Once
	stdPackages     []string
)

func completeStandardPackages(result map[string]string, prefix string) {
	stdPackagesOnce.Do(func() {
		cmd := exec.Command("go", "list", "std")
		output, err := cmd.Output()
		if err != nil {
			return
		}
		stdPackages = strings.Fields(string(bytes.TrimSpace(output)))
		slices.Sort(stdPackages)
	})

	for _, pkg := range stdPackages {
		if strings.HasPrefix(pkg, prefix) {
			result[pkg] = "std"
		}
	}
}

func completeLocalPackages(result map[string]string, prefix string) {
	if !strings.HasPrefix(prefix, "./") && prefix != "." {
		return
	}
	dir := "."
	if prefix != "." {
		dir = filepath.Clean(prefix)
	}
	baseDir := dir
	namePrefix := ""
	if !strings.HasSuffix(prefix, "/") {
		baseDir, namePrefix = filepath.Split(dir)
		if baseDir == "" {
			baseDir = "."
		}
	}
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() || ignoredPackageDir(entry.Name()) || !strings.HasPrefix(entry.Name(), namePrefix) {
			continue
		}
		p := filepath.ToSlash(filepath.Join(baseDir, entry.Name()))
		if p == "." {
			p = "./"
		} else if !strings.HasPrefix(p, "./") {
			p = "./" + p
		}
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		result[p] = "local"
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

// ModCache reads a Go module cache laid out like GOPATH/pkg/mod (or GOMODCACHE).
//
// See the Go module cache reference for the canonical cache description:
// https://go.dev/ref/mod#module-cache
//
//	<cache>                ::= <extract-tree> | "cache/download/" <download-tree>
//	<extract-tree>         ::= <escaped-module-path> "@" <escaped-version> "/" ...
//	<download-tree>        ::= <escaped-module-path> "/@v/" <version-file>
//	<version-file>         ::= "list" | <escaped-version> ".info" | ...
//	<package-dir>          ::= <escaped-module-path> "@" <escaped-version> ["/" <package-suffix>]
//	<package-suffix>       ::= <path-component> ["/" <package-suffix>]
//
// Module and version completion use both cache trees:
//
//   - Extracted module directories live directly under the cache root as
//     escaped module paths suffixed with @version.
//
//   - Download metadata lives under cache/download/<escaped-module>/@v. The
//     @v/list file contains newline-separated versions known to the go command.
type ModCache struct {
	fs  fs.ReadDirFS
	log func(format string, args ...any)
}

func GetModCachePath() string {
	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		gomodcache = filepath.Join(gopath, "pkg", "mod")
	}
	return gomodcache
}

func NewModCache() ModCache {
	return ModCache{
		fs:  os.DirFS(GetModCachePath()).(fs.ReadDirFS),
		log: func(format string, args ...any) {},
	}
}

func (m ModCache) CompleteModules(result map[string]string, prefix string) {
	module, version, hasVersion := strings.Cut(prefix, "@")
	escaped, err := escapePath(module)
	if err != nil {
		return
	}
	if hasVersion {
		m.completeVersion(result, module, version, escaped, "")
	} else {
		m.completeModule(result, module, escaped)
	}
}

func (m ModCache) CompletePackages(result map[string]string, prefix string) {
	pkg, version, hasVersion := strings.Cut(prefix, "@")
	escaped, err := escapePath(pkg)
	if err != nil {
		return
	}
	for i, ch := range escaped {
		if ch != filepath.Separator {
			continue
		}
		modpath, relpath := escaped[:i], escaped[i:]
		m.log("Complete package %q version %q modpath %q relpath %q", pkg, version, modpath, relpath)
		if hasVersion {
			m.completeVersion(result, pkg, version, modpath, relpath)
		} else {
			m.completeModulePackage(result, pkg, modpath, relpath)
		}
	}
	if hasVersion {
		m.completeVersion(result, pkg, version, escaped, "")
	} else {
		m.completeModule(result, pkg, escaped)
	}
}

func (m ModCache) completeModule(result map[string]string, pkg, modpath string) {
	parent, _ := path.Split(pkg)
	moddir, modprefix := filepath.Split(modpath)
	entries, err := m.fs.ReadDir(filepath.Clean(moddir))
	if err != nil {
		m.log("moddir %q error %v", moddir, err)
		return
	}
	m.log("Complete module package %q moddir %q modprefix %q", pkg, moddir, modprefix)
	for _, entry := range entries {
		name := entry.Name()
		m.log("package %q moddir %q prefix %q name %q", pkg, moddir, modprefix, name)
		if !entry.IsDir() || !strings.HasPrefix(name, modprefix) || moddir == "" && name == "cache" {
			continue
		}
		name, _, hasVersion := strings.Cut(name, "@")
		if name, err := unescapePath(name); err == nil {
			result[parent+name+"/"] = "cache"
			if hasVersion {
				result[parent+name+"@"] = "cache"
			}
		} else {
			m.log("cannot unescape %q error %v", name, err)
		}
	}
}

func (m ModCache) completeModulePackage(result map[string]string, pkg, modpath, relpath string) {
	moddir, modname := filepath.Split(modpath)
	entries, err := m.fs.ReadDir(filepath.Clean(moddir))
	if err != nil {
		m.log("moddir %q err %q", moddir, err)
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || !strings.HasPrefix(name, modname+"@") {
			continue
		}
		prefixpath := filepath.Join(moddir, name) + relpath
		m.log("package %q moddir %q modname %q name %q prefixpath %q", pkg, moddir, modname, name, prefixpath)
		m.completePackage(result, pkg, prefixpath)
		if !strings.HasSuffix(pkg, "/") {
			// m.completePackage(result, pkg+"/", prefixpath+string(filepath.Separator))
		}
	}
}

func (m ModCache) completePackage(result map[string]string, pkg, pkgpath string) {
	parent, _ := path.Split(pkg)
	parentDir, namePrefix := filepath.Split(pkgpath)
	entries, err := m.fs.ReadDir(filepath.Clean(parentDir))
	if err != nil {
		m.log("package %q directory %q error %v", pkg, parentDir, err)
		return
	}
	m.log("package %q parent %q directory %q prefix %q", pkg, parent, parentDir, namePrefix)
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || !strings.HasPrefix(name, namePrefix) || ignoredPackageDir(name) {
			continue
		}
		if pkgname, err := unescapePath(name); err == nil {
			result[parent+pkgname+"/"] = "cache"
		} else {
			m.log("cannot unescape %q error %v", name, err)
		}
	}
}

func (m ModCache) completeVersion(result map[string]string, pkg, versionPrefix, modpath, relpath string) {
	m.log("complete version for package %q version %q modpath %q", pkg, versionPrefix, modpath)
	for _, v := range []string{"latest", "patch"} {
		if strings.HasPrefix(v, versionPrefix) {
			result[pkg+"@"+v] = v
		}
	}
	if data, err := fs.ReadFile(m.fs, moduleVersionsPath(modpath)); err == nil {
		scan := bufio.NewScanner(strings.NewReader(string(data)))
		for scan.Scan() {
			if v := scan.Text(); strings.HasPrefix(v, versionPrefix) {
				if status, ok := m.getPacakgeStatus(modpath, relpath, v); ok {
					result[pkg+"@"+v] = status
				}
			}
		}
	}
}

func (m ModCache) getPacakgeStatus(modpath, relpath, version string) (string, bool) {
	if escaped, err := module.EscapeVersion(version); err == nil {
		if st, err := fs.Stat(m.fs, modpath+"@"+escaped+relpath); err == nil {
			return "cache", st.IsDir()
		}
		if _, err := fs.Stat(m.fs, modpath+"@"+escaped); err == nil {
			return "", false // module in cache but has no package
		}
	}
	return "", true
}

func escapePath(name string) (string, error) {
	escaped, err := module.EscapePath("x.x/" + name + "x")
	if err != nil {
		return "", err
	}
	return escaped[4 : len(escaped)-1], nil
}

func unescapePath(escaped string) (string, error) {
	unescaped, err := module.UnescapePath("x.x/" + escaped + "x")
	if err != nil {
		return "", err
	}
	return unescaped[4 : len(unescaped)-1], nil
}

func moduleVersionsPath(modpath string) string {
	return path.Join("cache", "download", modpath, "@v", "list")
}

func ignoredPackageDir(name string) bool {
	return name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}
