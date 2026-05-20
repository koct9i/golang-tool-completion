package gomodules

import (
	_ "embed"
	"strings"
)

//go:generate bash -c "curl https://goproxy.cn/stats/trends/last-30-days | jq -r '[.[].module_path] | sort | .[]' >popular.txt"

//go:embed popular.txt
var popular string

func CompletePopular(results map[string]string, prefix string) {
	if strings.Contains(prefix, "/") {
		for line := range strings.Lines(popular) {
			if strings.HasPrefix(line, prefix) {
				results[strings.TrimSuffix(line, "\n")] = "popular"
			}
		}
	}
}
