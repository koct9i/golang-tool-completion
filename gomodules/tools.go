package gomodules

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

//go:embed tools.txt
var tools string

func readTools() {
	userToolsPath, err := userToolsPath(false)
	if err != nil {
		return
	}
	//nolint:gosec //read
	userTools, err := os.ReadFile(userToolsPath)
	if err == nil {
		tools = string(userTools)
	}
}

func userToolsPath(mkdir bool) (string, error) {
	return userDataPath("tools.txt", mkdir)
}

func AddTools(packages, descriptions []string) error {
	userToolsPath, err := userToolsPath(true)
	if err != nil {
		return err
	}
	readTools()
	if len(descriptions) == 0 {
		descriptions = []string{""}
	}
	for i, pkg := range packages {
		description := descriptions[min(i, len(descriptions)-1)]
		tools += fmt.Sprintf("%s\t%s\n", pkg, description)
	}
	//nolint:gosec //0o644
	return os.WriteFile(userToolsPath, []byte(tools), 0o644)
}

func CompleteToolPackages(results map[string]string, prefix string) {
	readTools()
	completeKnownPackages(results, prefix, tools, "tool")
}

func completeKnownPackages(results map[string]string, prefix, packages, defaultDescription string) {
	if strings.HasPrefix(prefix, ".") {
		return
	}
	index := 0
	for line := range strings.Lines(packages) {
		index++
		if tail, found := strings.CutPrefix(line, prefix); found {
			tail, _ = strings.CutSuffix(tail, "\n")
			var desc string
			if tail, desc, found = strings.Cut(tail, "\t"); !found {
				desc = fmt.Sprintf("%s #%v", defaultDescription, index)
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
			pkg, _ := strings.CutSuffix(line, "\n")
			pkg, desc, found := strings.Cut(pkg, "\t")
			if !found {
				desc = fmt.Sprintf("%s #%v", defaultDescription, index)
			}
			_, tail, _ := strings.Cut(pkg, "/")
			if strings.Contains("/"+tail, prefix) {
				results[pkg+"@"] = desc
			}
		}
	}
}
