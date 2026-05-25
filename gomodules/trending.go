package gomodules

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:generate bash -c "curl https://goproxy.cn/stats/trends/last-30-days | jq -r '[.[].module_path] | sort | .[]' >trending.txt"

//go:embed trending.txt
var trending string

func readTrending() {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	userTrendingPath := filepath.Join(userConfigDir, "golang-tool-completion", "trending.txt")
	//nolint:gosec //read
	userTrending, err := os.ReadFile(userTrendingPath)
	if err == nil {
		trending = string(userTrending)
	}
}

func DisableTrending() error {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	userTrendingPath := filepath.Join(userConfigDir, "golang-tool-completion", "trending.txt")
	if _, err := fmt.Fprintf(os.Stderr, "Disabling completion for trending modules by creating empty file: %v\n", userTrendingPath); err != nil {
		return err
	}
	//nolint:gosec //0o644
	f, err := os.OpenFile(userTrendingPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

func CompleteTrending(results map[string]string, prefix string) {
	readTrending()
	for line := range strings.Lines(trending) {
		if tail, found := strings.CutPrefix(line, prefix); found {
			tail, _ = strings.CutSuffix(tail, "\n")
			if tail, _, found = strings.Cut(tail, "/"); found {
				results[prefix+tail+"/"] = "trending"
			} else {
				results[prefix+tail+"@"] = "trending"
			}
		}
	}
}
