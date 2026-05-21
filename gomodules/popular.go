package gomodules

import (
	_ "embed"
	"strings"
)

//go:generate bash -c "curl https://goproxy.cn/stats/trends/last-30-days | jq -r '[.[].module_path] | sort | .[]' >popular.txt"

//go:embed popular.txt
var popular string

func CompletePopular(results map[string]string, prefix string) {
	for line := range strings.Lines(popular) {
		if tail, found := strings.CutPrefix(line, prefix); found {
			tail, _ = strings.CutSuffix(tail, "\n")
			if tail, _, found = strings.Cut(tail, "/"); found {
				results[prefix+tail+"/"] = "popular"
			} else {
				results[prefix+tail+"@"] = "popular"
			}
		}
	}
}
