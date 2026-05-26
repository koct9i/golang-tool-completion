package gomodules

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:generate bash -c "curl -s https://goproxy.cn/stats/trends/last-30-days | jq -r '.[] | [ .module_path, (.description // .module_description // .repo_description // \"\") ] | @tsv' | sort >trending.txt"

//go:embed trending.txt
var embeddedTrending string

var trending = embeddedTrending

func readTrending() {
	userTrendingPath, err := userTrendingPath()
	if err != nil {
		return
	}
	//nolint:gosec //read
	userTrending, err := os.ReadFile(userTrendingPath)
	if err == nil {
		trending = string(userTrending)
	}
}

func userTrendingPath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userConfigDir, "golang-tool-completion", "trending.txt"), nil
}

func DisableTrending() error {
	userTrendingPath, err := userTrendingPath()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stderr, "Disabling completion for trending modules by creating empty file: %v\n", userTrendingPath); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(userTrendingPath), 0o755); err != nil {
		return err
	}
	//nolint:gosec //0o644
	f, err := os.OpenFile(userTrendingPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

func AddTrending() error {
	userTrendingPath, err := userTrendingPath()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stderr, "Adding trending modules into config: %v\n", userTrendingPath); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(userTrendingPath), 0o755); err != nil {
		return err
	}
	//nolint:gosec //0o644
	if err := os.WriteFile(userTrendingPath, []byte(embeddedTrending), 0o644); err != nil {
		return err
	}
	return nil
}

func parseTrendingLine(line string) (modulePath string, description string, ok bool) {
	line = strings.TrimSuffix(line, "\n")
	if line == "" {
		return "", "", false
	}
	modulePath, description, _ = strings.Cut(line, "\t")
	if modulePath == "" {
		return "", "", false
	}
	if description == "" {
		description = "trending"
	}
	return modulePath, description, true
}

func CompleteTrending(results map[string]string, prefix string) {
	readTrending()
	for line := range strings.Lines(trending) {
		modulePath, description, found := parseTrendingLine(line)
		if !found {
			continue
		}
		if tail, found := strings.CutPrefix(modulePath, prefix); found {
			if tail, _, found = strings.Cut(tail, "/"); found {
				results[prefix+tail+"/"] = "trending"
			} else {
				results[prefix+tail+"@"] = description
			}
		}
	}
}
