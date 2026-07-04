package gomodules

import (
	"bufio"
	"context"
	"io/fs"
	"iter"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

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
	dir string
	fs  fs.ReadDirFS
	log func(format string, args ...any)
}

func GetModCachePath() string {
	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}
			gopath = filepath.Join(home, "go")
		}
		gomodcache = filepath.Join(gopath, "pkg", "mod")
	}
	return gomodcache
}

var Log func(format string, args ...any) = func(format string, args ...any) {}

func NewModCache() ModCache {
	dir := GetModCachePath()
	//nolint:errcheck //fs.ReadDirFS
	fs := os.DirFS(dir).(fs.ReadDirFS)
	return ModCache{
		dir: dir,
		fs:  fs,
		log: Log,
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

func (m ModCache) iterPackageVersions(prefix string) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		escaped, err := escapePath(prefix)
		if err != nil {
			return
		}
		n := len(escaped)
		for i := range n + 1 {
			if i < n && escaped[i] != filepath.Separator {
				continue
			}
			modpath, relpath := escaped[:min(i, n)], escaped[min(i, n):]
			moddir, modname := filepath.Split(modpath)
			entries, err := m.fs.ReadDir(filepath.Clean(moddir))
			if err != nil {
				m.log("readdir %q err %q", moddir, err)
				continue
			}
			for _, entry := range entries {
				entryname := entry.Name()
				name, version, hasVersion := strings.Cut(entryname, "@")
				if !entry.IsDir() || !hasVersion || name != modname {
					continue
				}
				if v, err := module.UnescapeVersion(version); err == nil {
					version = v
				}
				packagepath := filepath.Join(moddir, entryname) + relpath
				m.log("package %q path %q version %q", prefix, packagepath, version)
				if !yield(packagepath, version) {
					break
				}
			}
		}
	}
}

func (m ModCache) CompleteMainPackages(ctx context.Context, result map[string]string, prefix string) {
	if strings.HasPrefix(prefix, ".") {
		return
	}
	pkgPrefix, versionPrefix, hasVersion := strings.Cut(prefix, "@")
	var bestDir, bestVersion string
	for pkgpath, version := range m.iterPackageVersions(pkgPrefix) {
		if hasVersion {
			if !strings.HasPrefix(version, versionPrefix) {
				continue
			}
			if st, err := fs.Stat(m.fs, pkgpath); err == nil && st.IsDir() {
				result[pkgPrefix+"@"+version] = "cache"
			} else {
				m.log("stat %q: %v", pkgpath, err)
			}
		} else if bestDir == "" || semver.Compare(version, bestVersion) > 0 {
			pkgdir := filepath.Dir(pkgpath)
			if st, err := fs.Stat(m.fs, pkgdir); err == nil && st.IsDir() {
				bestDir, bestVersion = pkgdir, version
			}
		}
	}
	if hasVersion {
		for _, v := range []string{"latest", "patch"} {
			if strings.HasPrefix(v, versionPrefix) {
				result[pkgPrefix+"@"+v] = v
			}
		}
		return
	}
	if bestDir == "" || m.dir == "" {
		return
	}
	currentDirectory := ChangeDirectory
	ChangeDirectory = filepath.Join(m.dir, bestDir)
	output, err := runGo(ctx, "list", "-f", `{{if eq .Name "main"}}{{.ImportPath}}{{end}}`, "./...")
	ChangeDirectory = currentDirectory
	if err != nil {
		m.log("go list main packages in %q error %v", bestDir, err)
		return
	}
	for line := range strings.Lines(string(output)) {
		pkg := strings.TrimSpace(line)
		if pkg == "" || !strings.HasPrefix(pkg, prefix) {
			continue
		}
		result[pkg+"@"] = "cache"
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

func (m ModCache) CompleteUsedModules(ctx context.Context, result map[string]string, prefix string) {
	pkg, version, hasVersion := strings.Cut(prefix, "@")
	output, err := runGo(ctx, "list", "-m", "all")
	if err != nil {
		return
	}
	for line := range strings.Lines(string(output)) {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		mod := fields[0]
		ver := ""
		hasVer := len(fields) > 1
		if hasVer {
			ver = fields[1]
		}
		if hasVersion {
			if mod != pkg {
				continue
			}
			if escaped, err := escapePath(mod); err == nil {
				m.completeVersion(result, mod, version, escaped, "")
			}
			if strings.HasPrefix(ver, version) {
				result[pkg+"@"+ver] = "used"
			}
		} else if hasVer && isMatchingPackage(mod, pkg) {
			result[mod+"@"] = "used"
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
				if status, ok := m.getPackageStatus(modpath, relpath, v); ok {
					result[pkg+"@"+v] = status
				}
			}
		}
	}
}

func isMatchingPackage(pkg, input string) bool {
	return strings.HasPrefix(pkg, input) || len(input) >= MinSubstringLen && strings.Contains(pkg, input)
}

func (m ModCache) getPackageStatus(modpath, relpath, version string) (string, bool) {
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
