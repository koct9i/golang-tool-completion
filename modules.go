package main

import (
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
	escaped, err := module.EscapePath("x.x/" + modpath + "x")
	if err != nil {
		return nil
	}
	escaped = escaped[4 : len(escaped)-1]
	if hasVersion {
		escapedVersion, err := module.EscapeVersion(version)
		if err != nil {
			escapedVersion = version
		}
		escaped = escaped + "@" + escapedVersion
	}

	cache := GoModCacheFS()
	enties, err := cache.ReadDir(path.Dir(escaped + "x"))
	if err != nil {
		return nil
	}
	dir, namePrefix := path.Split(escaped)
	result := make(map[string]string, min(len(enties), 1000))
	suggestNext := ""
	for _, entry := range enties {
		if name := entry.Name(); entry.IsDir() && strings.HasPrefix(name, namePrefix) {
			name, ver, hasVer := strings.Cut(name, "@")

			suggest, err := module.UnescapePath(dir + name)
			if err != nil {
				continue
			}

			if hasVer && hasVersion {
				ver, err := module.UnescapeVersion(ver)
				if err != nil {
					continue
				}
				suggest = suggest + "@" + ver
			}

			result[suggest] = ""

			if suggestNext == "" {
				if !hasVer {
					suggestNext = suggest + "/"
				} else if !hasVersion {
					suggestNext = suggest + "@"
				}
			}
		}
	}
	if len(result) == 1 && suggestNext != "" {
		result = map[string]string{suggestNext: ""}
	}
	return result
}
