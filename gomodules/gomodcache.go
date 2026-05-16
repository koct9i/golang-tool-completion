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

type ModCache struct {
	fs fs.ReadDirFS
}

func NewModCache() ModCache {
	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		gomodcache = filepath.Join(gopath, "pkg", "mod")
	}
	return ModCache{
		fs: os.DirFS(gomodcache).(fs.ReadDirFS),
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
	entries, _ := m.fs.ReadDir(path.Dir(escaped+"x"))
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || name == "@v" || !strings.HasPrefix(name, namePrefix) {
			continue
		}
		base, ver, hasVersion := strings.Cut(name, "@")
		suggest, err := module.UnescapePath(dir + base)
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

func (m ModCache) completeModuleVersions(result map[string]string, modpath, prefix string) {
	for _, v := range []string{"latest", "patch", "none"} {
		if strings.HasPrefix(v, prefix) {
			result[modpath+"@"+v] = ""
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
				result[modpath+"@"+v] = ""
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
				result[modpath+"@"+v] = ""
			}
		}
	}
	entries, _ := m.fs.ReadDir(versionDir)
	for _, entry := range entries {
		if v, ok := strings.CutSuffix(entry.Name(), ".info"); ok && strings.HasPrefix(v, prefix) {
			if v, err := module.UnescapeVersion(v); err == nil {
				result[modpath+"@"+v] = ""
			}
		}
	}
}

func escapePathPrefix(prefix string) (string, bool) {
	escaped, err := module.EscapePath("x.x/" + prefix + "x")
	if err != nil {
		return "", false
	}
	return escaped[4 : len(escaped)-1], true
}
