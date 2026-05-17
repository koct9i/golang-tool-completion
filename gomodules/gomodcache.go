package gomodules

import (
	"bufio"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"
)

func CompleteModules(prefix string) map[string]string {
	return NewModCache().CompleteModules(prefix)
}

func CompletePackages(prefix string) map[string]string {
	return NewModCache().CompletePackages(prefix)
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

func (m ModCache) CompleteModules(prefix string) map[string]string {
	module, version, hasVersion := strings.Cut(prefix, "@")
	escaped, err := escapePath(module)
	if err != nil {
		return nil
	}
	result := map[string]string{}
	if hasVersion {
		m.completeVersion(result, module, version, escaped)
	} else {
		m.completeModule(result, module, escaped)
	}
	return result
}

func (m ModCache) CompletePackages(prefix string) map[string]string {
	pkg, version, hasVersion := strings.Cut(prefix, "@")
	escaped, err := escapePath(pkg)
	if err != nil {
		return nil
	}
	result := map[string]string{}
	for i, ch := range escaped {
		if ch == filepath.Separator {
			modpath, relpath := escaped[:i], escaped[i+1:]
			m.log("Complete package %q version %q modpath %q", pkg, version, modpath)
			if hasVersion {
				m.completeVersion(result, pkg, version, modpath)
			} else {
				m.completeModulePackage(result, pkg, modpath, relpath)
			}
		}
	}
	if hasVersion {
		m.completeVersion(result, pkg, version, escaped)
	} else {
		m.completeModule(result, pkg, escaped)
	}
	return result
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
		if !entry.IsDir() || !strings.HasPrefix(name, modprefix) {
			continue
		}
		name, _, hasVersion := strings.Cut(name, "@")
		if name, err := unescapePath(name); err == nil {
			result[parent+name+"/"] = ""
			if hasVersion {
				result[parent+name+"@"] = ""
			}
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
		m.log("moddir %q modname %q name %q", moddir, modname, name)
		if !entry.IsDir() || !strings.HasPrefix(name, modname+"@") {
			continue
		}
		prefixpath := filepath.Join(moddir, name, relpath)
		m.completePackage(result, pkg, prefixpath)
		if !strings.HasSuffix(pkg, "/") {
			m.completePackage(result, pkg+"/", prefixpath+string(filepath.Separator))
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
	m.log("package %q directory %q", pkg, parentDir)
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || !strings.HasPrefix(name, namePrefix) || ignoredPackageDir(name) {
			continue
		}
		if pkgname, err := unescapePath(name); err == nil {
			result[parent+pkgname] = ""
		}
	}
}

func (m ModCache) completeVersion(result map[string]string, pkg, versionPrefix, modpath string) {
	m.log("complete version for package %q version %q modpath %q", pkg, versionPrefix, modpath)
	for _, v := range []string{"latest", "patch"} {
		if strings.HasPrefix(v, versionPrefix) {
			result[pkg+"@"+v] = ""
		}
	}
	if data, err := fs.ReadFile(m.fs, moduleVersionsPath(modpath)); err == nil {
		scan := bufio.NewScanner(strings.NewReader(string(data)))
		for scan.Scan() {
			if v := scan.Text(); strings.HasPrefix(v, versionPrefix) {
				result[pkg+"@"+v] = ""
			}
		}
	}
}

func escapePath(name string) (string, error) {
	escaped, err := module.EscapePath("x.x/" + name + "x")
	if err != nil {
		return "", err
	}
	return escaped[4 : len(escaped)-1], nil
}

func unescapePath(escaped string) (string, error) {
	return module.UnescapePath(escaped)
}

func moduleVersionsPath(modpath string) string {
	return path.Join("cache", "download", modpath, "@v", "list")
}

func ignoredPackageDir(name string) bool {
	return name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}
