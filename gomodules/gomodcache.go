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

func GoModCacheFS() fs.ReadDirFS {
	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		gomodcache = filepath.Join(gopath, "pkg", "mod")
	}
	return os.DirFS(gomodcache).(fs.ReadDirFS)
}

func CompleteModules(prefix string) map[string]string {
	modpath, version, hasVersion := strings.Cut(prefix, "@")
	result := map[string]string{}
	if hasVersion {
		completeModuleVersions(result, modpath, version)
	} else {
		completeModulePaths(result, modpath)
	}
	return result
}

func completeModulePaths(result map[string]string, prefix string) {
	escaped, ok := escapePathPrefix(prefix)
	if !ok {
		return
	}
	completePathDir(result, "", escaped, false)
	completePathDir(result, "cache/download", escaped, true)
}

func completeModuleVersions(result map[string]string, modpath, prefix string) {
	for _, v := range []string{"latest", "patch", "none"} {
		if strings.HasPrefix(v, prefix) {
			result[modpath+"@"+v] = ""
		}
	}

	escaped, ok := escapePathPrefix(modpath + "@" + prefix)
	if ok {
		dir, namePrefix := path.Split(escaped)
		for _, entry := range readDir(path.Dir(dir + "x")) {
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
	if data, err := fs.ReadFile(GoModCacheFS(), path.Join(versionDir, "list")); err == nil {
		scan := bufio.NewScanner(strings.NewReader(string(data)))
		for scan.Scan() {
			if v := scan.Text(); strings.HasPrefix(v, prefix) {
				result[modpath+"@"+v] = ""
			}
		}
	}
	for _, entry := range readDir(versionDir) {
		if v, ok := strings.CutSuffix(entry.Name(), ".info"); ok && strings.HasPrefix(v, prefix) {
			if v, err := module.UnescapeVersion(v); err == nil {
				result[modpath+"@"+v] = ""
			}
		}
	}
}

func completePathDir(result map[string]string, root, escaped string, download bool) {
	dir, namePrefix := path.Split(escaped)
	for _, entry := range readDir(path.Join(root, path.Dir(escaped+"x"))) {
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
		if download && hasList(path.Join(root, dir, name, "@v", "list")) {
			result[suggest+"@"] = ""
		} else {
			result[suggest+"/"] = ""
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

func readDir(name string) []fs.DirEntry {
	entries, _ := GoModCacheFS().ReadDir(name)
	return entries
}

func hasList(name string) bool {
	_, err := fs.Stat(GoModCacheFS(), name)
	return err == nil
}
