package gomodules

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

//go:generate bash -c "curl -s https://goproxy.cn/stats/trends/last-30-days | jq -r 'sort_by(.download_count)|reverse|.[].module_path' >trending.txt"

//go:embed trending.txt
var trending string

func readTrending() {
	userTrendingPath, err := userTrendingPath(false)
	if err != nil {
		return
	}
	//nolint:gosec //read
	userTrending, err := os.ReadFile(userTrendingPath)
	if err == nil {
		trending = string(userTrending)
	}
}

func userDataPath(filename string, mkdir bool) (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(userConfigDir, "golang-tool-completion")
	if mkdir {
		st, err := os.Stat(configDir)
		if err != nil && os.IsNotExist(err) {
			//nolint:gosec //0o755
			err = os.Mkdir(configDir, 0o755)
		} else if err == nil && !st.IsDir() {
			err = fmt.Errorf("not a directory: %v", configDir)
		}
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(configDir, filename), nil
}

func userTrendingPath(mkdir bool) (string, error) {
	return userDataPath("trending.txt", mkdir)
}

func DisableTrending() error {
	userTrendingPath, err := userTrendingPath(true)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stderr, "Disabling completion for trending modules by placing empty file: %v\n", userTrendingPath); err != nil {
		return err
	}
	trending = ""
	//nolint:gosec //0o644
	return os.WriteFile(userTrendingPath, nil, 0o644)
}

func AddTrending(modules, descriptions []string) error {
	userTrendingPath, err := userTrendingPath(true)
	if err != nil {
		return err
	}
	readTrending()
	if len(descriptions) == 0 {
		descriptions = []string{""}
	}
	for i, module := range modules {
		description := descriptions[min(i, len(descriptions)-1)]
		trending += fmt.Sprintf("%s\t%s\n", module, description)
	}
	//nolint:gosec //0o644
	return os.WriteFile(userTrendingPath, []byte(trending), 0o644)
}

func CompleteTrending(results map[string]string, prefix string) {
	readTrending()
	completeKnownPackages(results, prefix, trending, "trending", strings.Count(prefix, "/") < 2)
}

func completeKnownPackages(results map[string]string, prefix, packages, defaultDescription string, shallow bool) {
	index := 0
	for line := range strings.Lines(packages) {
		index++
		if tail, found := strings.CutPrefix(line, prefix); found {
			tail, _ = strings.CutSuffix(tail, "\n")
			var desc string
			if tail, desc, found = cutDesc(tail); !found {
				desc = fmt.Sprintf("%s #%v", defaultDescription, index)
			}
			if word, _, found := strings.Cut(tail, "/"); found && shallow {
				if _, found := results[prefix+word+"/"]; !found {
					results[prefix+word+"/"] = desc
				}
			} else {
				if !strings.Contains(tail, "@") {
					tail += "@"
				}
				results[prefix+tail] = desc
			}
		}
		if len(prefix) >= MinSubstringLen && strings.Contains(line, prefix) {
			pkg, _ := strings.CutSuffix(line, "\n")
			pkg, desc, found := cutDesc(pkg)
			if !found {
				desc = fmt.Sprintf("%s #%v", defaultDescription, index)
			}
			_, tail, _ := strings.Cut(pkg, "/")
			if strings.Contains("/"+tail, prefix) {
				if !strings.Contains(pkg, "@") {
					pkg += "@"
				}
				results[pkg] = desc
			}
		}
	}
}

func cutDesc(s string) (pkg, desc string, found bool) {
	if i := strings.IndexFunc(s, unicode.IsSpace); i >= 0 {
		return s[:i], strings.TrimSpace(s[i:]), true
	}
	return s, "", false
}
