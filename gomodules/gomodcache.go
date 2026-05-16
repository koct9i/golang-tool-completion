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

type ModCache struct {
	fs fs.ReadDirFS
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
		fs: os.DirFS(GetModCachePath()).(fs.ReadDirFS),
	}
}

func (m ModCache) CompleteModules(prefix string) map[string]string {
	modpath, version, hasVersion := strings.Cut(prefix, "@")
	result := map[string]string{}
	if hasVersion {
		m.completeModuleVersions(result, modpath, version)
	} else {
		m.completeModulePaths(result, modpath)
	}
	return result
}

func (m ModCache) CompletePackages(prefix string) map[string]string {
	pkgpath, version, hasVersion := strings.Cut(prefix, "@")
	result := map[string]string{}
	if hasVersion {
		m.completePackageVersions(result, pkgpath, version)
	} else {
		m.completePackageModulePathPrefixes(result, pkgpath)
		m.completePackagePaths(result, pkgpath)
	}
	return result
}

func (m ModCache) completeModulePaths(result map[string]string, prefix string) {
	escaped, ok := escapePathPrefix(prefix)
	if !ok {
		return
	}
	m.completePathDir(result, escaped, false)
	m.completePathDir(result, path.Join("cache/download", escaped), true)
}

func (m ModCache) completePathDir(result map[string]string, escaped string, download bool) {
	dir, namePrefix := path.Split(escaped)
	entries, _ := m.fs.ReadDir(path.Dir(escaped + "x"))
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || name == "@v" || !strings.HasPrefix(name, namePrefix) {
			continue
		}
		base, ver, hasVersion := strings.Cut(name, "@")
		suggestPath := dir + base
		if download {
			suggestPath = strings.TrimPrefix(suggestPath, "cache/download/")
		}
		suggest, err := module.UnescapePath(suggestPath)
		if err != nil {
			continue
		}
		if hasVersion {
			if _, err = module.UnescapeVersion(ver); err == nil {
				result[suggest+"@"] = ""
			}
			continue
		}
		if !download {
			result[suggest+"/"] = ""
		} else if _, err := fs.Stat(m.fs, path.Join(dir, name, "@v", "list")); err == nil {
			result[suggest+"@"] = ""
		}
	}
}

func (m ModCache) completePackageModulePathPrefixes(result map[string]string, prefix string) {
	escaped, ok := escapePathPrefix(prefix)
	if !ok {
		return
	}
	m.completePackageModulePathDir(result, escaped)
	m.completePackageModulePathDir(result, path.Join("cache/download", escaped))
}

func (m ModCache) completePackageModulePathDir(result map[string]string, escaped string) {
	dir, namePrefix := path.Split(escaped)
	entries, _ := m.fs.ReadDir(path.Dir(escaped + "x"))
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || name == "@v" || !strings.HasPrefix(name, namePrefix) {
			continue
		}
		base, _, hasVersion := strings.Cut(name, "@")
		if hasVersion {
			continue
		}
		suggestPath := dir + base
		suggestPath = strings.TrimPrefix(suggestPath, "cache/download/")
		suggest, err := module.UnescapePath(suggestPath)
		if err != nil {
			continue
		}
		result[suggest+"/"] = ""
	}
}

func (m ModCache) completePackagePaths(result map[string]string, prefix string) {
	escaped, ok := escapePathPrefix(prefix)
	if !ok {
		return
	}
	m.walkModuleDirs(func(mod module.Version, dir string) bool {
		if mod.Path == "" {
			return true
		}
		suffixPrefix, ok := packageSuffixPrefix(mod.Path, prefix)
		if !ok {
			return true
		}
		m.completePackagesInModule(result, mod.Path, dir, suffixPrefix)
		return true
	})

	// A package prefix may include more path elements than the module path. In that
	// case, module-directory discovery cannot narrow the search before it knows the
	// module boundary. Use the escaped package prefix to find already-extracted
	// module directories whose names begin at the same path location, including
	// semantic import version suffixes such as /v2.
	dir, namePrefix := path.Split(escaped)
	entries, _ := m.fs.ReadDir(path.Dir(escaped + "x"))
	for _, entry := range entries {
		name := entry.Name()
		base, ver, hasVersion := strings.Cut(name, "@")
		if !entry.IsDir() || !hasVersion || !strings.HasPrefix(base+"@"+ver, namePrefix) {
			continue
		}
		modpath, err := module.UnescapePath(dir + base)
		if err != nil {
			continue
		}
		if _, err := module.UnescapeVersion(ver); err != nil {
			continue
		}
		m.completePackagesInModule(result, modpath, path.Join(dir, name), "")
	}
}

func (m ModCache) completePackagesInModule(result map[string]string, modpath, dir, prefix string) {
	parent, namePrefix := path.Split(prefix)
	start := path.Join(dir, parent)
	fs.WalkDir(m.fs, start, func(pkgDir string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if pkgDir != start && ignoredPackageDir(d.Name()) {
			return fs.SkipDir
		}
		if parent == "" && pkgDir != start {
			rel := strings.TrimPrefix(strings.TrimPrefix(pkgDir, start), "/")
			first, _, _ := strings.Cut(rel, "/")
			if !strings.HasPrefix(first, namePrefix) {
				return fs.SkipDir
			}
		}
		entries, err := m.fs.ReadDir(pkgDir)
		if err != nil {
			return nil
		}
		hasPackage := false
		hasSubdir := false
		for _, entry := range entries {
			if entry.IsDir() {
				if !ignoredPackageDir(entry.Name()) {
					hasSubdir = true
				}
			} else if strings.HasSuffix(entry.Name(), ".go") {
				hasPackage = true
			}
		}
		pkgpath := modpath
		if suffix := strings.TrimPrefix(strings.TrimPrefix(pkgDir, dir), "/"); suffix != "" {
			pkgpath = modpath + "/" + suffix
		}
		if !strings.HasPrefix(path.Base(pkgpath), namePrefix) && pkgDir == start {
			return nil
		}
		if hasPackage {
			result[pkgpath] = ""
		}
		if hasSubdir {
			result[pkgpath+"/"] = ""
		}
		return nil
	})
}

func (m ModCache) completePackageVersions(result map[string]string, pkgpath, prefix string) {
	for _, v := range []string{"latest", "patch", "none"} {
		if strings.HasPrefix(v, prefix) {
			result[pkgpath+"@"+v] = ""
		}
	}
	m.walkModuleDirs(func(mod module.Version, dir string) bool {
		if mod.Path == "" {
			return true
		}
		suffix, ok := packageSuffix(mod.Path, pkgpath)
		if !ok || !m.dirHasPackage(path.Join(dir, suffix)) {
			return true
		}
		m.completeVersions(result, pkgpath, mod.Path, prefix)
		return true
	})
}

func (m ModCache) completeVersions(result map[string]string, suggestPath, modpath, prefix string) {
	for _, v := range []string{"latest", "patch", "none"} {
		if strings.HasPrefix(v, prefix) {
			result[suggestPath+"@"+v] = ""
		}
	}

	escaped, ok := escapePathPrefix(modpath + "@" + prefix)
	if ok {
		dir, namePrefix := path.Split(escaped)
		entries, _ := m.fs.ReadDir(path.Dir(dir + "x"))
		for _, entry := range entries {
			name, v, found := strings.Cut(entry.Name(), "@")
			if !entry.IsDir() || !found || !strings.HasPrefix(name+"@"+v, namePrefix) {
				continue
			}
			if v, err := module.UnescapeVersion(v); err == nil {
				result[suggestPath+"@"+v] = ""
			}
		}
	}

	escapedMod, err := module.EscapePath(modpath)
	if err != nil {
		return
	}
	versionDir := path.Join("cache/download", escapedMod, "@v")
	if data, err := fs.ReadFile(m.fs, path.Join(versionDir, "list")); err == nil {
		scan := bufio.NewScanner(strings.NewReader(string(data)))
		for scan.Scan() {
			if v := scan.Text(); strings.HasPrefix(v, prefix) {
				result[suggestPath+"@"+v] = ""
			}
		}
	}
	entries, _ := m.fs.ReadDir(versionDir)
	for _, entry := range entries {
		if v, ok := strings.CutSuffix(entry.Name(), ".info"); ok && strings.HasPrefix(v, prefix) {
			if v, err := module.UnescapeVersion(v); err == nil {
				result[suggestPath+"@"+v] = ""
			}
		}
	}
}

func (m ModCache) walkModuleDirs(fn func(module.Version, string) bool) {
	fs.WalkDir(m.fs, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if name == "." {
			return nil
		}
		if name == "cache" {
			return fs.SkipDir
		}
		base, ver, ok := strings.Cut(path.Base(name), "@")
		if !ok {
			return nil
		}
		modpath, err := module.UnescapePath(path.Join(path.Dir(name), base))
		if err != nil {
			return fs.SkipDir
		}
		version, err := module.UnescapeVersion(ver)
		if err != nil {
			return fs.SkipDir
		}
		if !fn(module.Version{Path: modpath, Version: version}, name) {
			return fs.SkipAll
		}
		return fs.SkipDir
	})
}

func packageSuffixPrefix(modpath, prefix string) (string, bool) {
	if prefix == modpath || strings.HasPrefix(modpath, prefix) {
		return "", true
	}
	if after, ok := strings.CutPrefix(prefix, modpath+"/"); ok {
		return after, true
	}
	return "", false
}

func packageSuffix(modpath, pkgpath string) (string, bool) {
	if pkgpath == modpath {
		return "", true
	}
	if after, ok := strings.CutPrefix(pkgpath, modpath+"/"); ok {
		return after, true
	}
	return "", false
}

func (m ModCache) dirHasPackage(dir string) bool {
	entries, err := m.fs.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true
		}
	}
	return false
}

func ignoredPackageDir(name string) bool {
	return name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

func (m ModCache) completeModuleVersions(result map[string]string, modpath, prefix string) {
	m.completeVersions(result, modpath, modpath, prefix)
}

func escapePathPrefix(prefix string) (string, bool) {
	escaped, err := module.EscapePath("x.x/" + prefix + "x")
	if err != nil {
		return "", false
	}
	return escaped[4 : len(escaped)-1], true
}
