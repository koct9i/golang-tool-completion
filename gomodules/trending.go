package gomodules

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func userTrendingPath(mkdir bool) (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	if mkdir {
		configDir := filepath.Join(userConfigDir, "golang-tool-completion")
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
	return filepath.Join(userConfigDir, "golang-tool-completion", "trending.txt"), nil
}

func DisableTrending() error {
	userTrendingPath, err := userTrendingPath(true)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stderr, "Disabling completion for trending modules by writing empty file: %v\n", userTrendingPath); err != nil {
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
	index := 0
	for line := range strings.Lines(trending) {
		index++
		if tail, found := strings.CutPrefix(line, prefix); found {
			tail, _ = strings.CutSuffix(tail, "\n")
			var desc string
			if tail, desc, found = strings.Cut(tail, "\t"); !found {
				desc = fmt.Sprintf("trending #%v", index)
			}
			if tail, _, found = strings.Cut(tail, "/"); found {
				if _, found := results[prefix+tail+"/"]; !found {
					results[prefix+tail+"/"] = desc
				}
			} else {
				results[prefix+tail+"@"] = desc
			}
		}
		if len(prefix) >= 3 && strings.Contains(line, prefix) {
			module, _ := strings.CutSuffix(line, "\n")
			module, desc, found := strings.Cut(module, "\t")
			if !found {
				desc = fmt.Sprintf("trending #%v", index)
			}
			_, tail, _ := strings.Cut(module, "/")
			if strings.Contains("/"+tail, prefix) {
				results[module+"@"] = desc
			}
		}
	}
}
