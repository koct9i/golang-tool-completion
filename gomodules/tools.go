package gomodules

import (
	_ "embed"
	"fmt"
	"os"
)

// Well-known installable Go tools are curated from project home pages and these indexes:
// https://github.com/avelino/awesome-go#go-tools
// https://github.com/avelino/awesome-go#devops-tools
// https://github.com/avelino/awesome-go#database-tools
// https://github.com/avelino/awesome-go#continuous-integration
// https://github.com/avelino/awesome-go#utilities
// https://github.com/gobuild/awesome-go-tools
// https://github.com/akavel/awesome-go#code-analysis
// https://github.com/akavel/awesome-go#command-line
// https://github.com/akavel/awesome-go#testing
// https://go.dev/wiki/CodeTools
// https://pkg.go.dev/golang.org/x/tools/cmd
// https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
// https://pkg.go.dev/golang.org/x/perf/cmd/benchstat

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

func DisableTools() error {
	userToolsPath, err := userToolsPath(true)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stderr, "Disabling completion for tools by placing empty file: %v\n", userToolsPath); err != nil {
		return err
	}
	tools = ""
	//nolint:gosec //0o644
	return os.WriteFile(userToolsPath, nil, 0o644)
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
	completeKnownPackages(results, prefix, tools, "tool", false)
}
