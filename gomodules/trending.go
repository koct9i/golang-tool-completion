package gomodules

import (
	_ "embed"
	"strings"
)

//go:generate bash -c "curl https://goproxy.cn/stats/trends/last-30-days | jq -r '[.[].module_path] | sort | .[]' >trending.txt"

//go:embed trending.txt
var trending string

func CompleteTrending(results map[string]string, prefix string) {
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
