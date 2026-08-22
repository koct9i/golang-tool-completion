package gomodules

import (
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"golang.org/x/mod/module"
)

func TestTrendingModules(t *testing.T) {
	for line := range strings.Lines(trending) {
		line, _ = strings.CutSuffix(line, "\n")
		mod, _, _ := strings.Cut(line, "\t")
		if err := module.CheckPath(mod); err != nil {
			t.Errorf("Trending %q %v", mod, err)
		}
	}

	current := trending
	t.Cleanup(func() {
		trending = current
	})
	trending += "example.com/mod\tExample module\n"

	for _, tt := range []struct {
		prefix string
		want   []string
		desc   []string
	}{
		{".", nil, nil},
		{"./", nil, nil},
		{"x/sync", []string{"golang.org/x/sync@"}, []string{"trending"}},
		{"golang.org/x/syn", []string{"golang.org/x/sync@"}, []string{"trending"}},
		{"example.co/", nil, nil},
		{"example.co", []string{"example.com/"}, []string{"Example module"}},
		{"example.com/", []string{"example.com/mod@"}, []string{"Example module"}},
	} {
		t.Run(tt.prefix, func(t *testing.T) {
			result := map[string]string{}
			CompleteTrending(result, tt.prefix)
			got := slices.Sorted(maps.Keys(result))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("CompleteTrending(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
			for i, mod := range tt.want {
				if desc := result[mod]; !strings.HasPrefix(desc, tt.desc[i]) {
					t.Fatalf("CompleteTrending(%q) %v = %v, want %v", tt.prefix, mod, desc, tt.desc[i])
				}
			}
		})
	}
}

func TestToolPackages(t *testing.T) {
	current := tools
	t.Cleanup(func() {
		tools = current
	})

	for line := range strings.Lines(tools) {
		line, _ = strings.CutSuffix(line, "\n")
		pkg, _, _ := strings.Cut(line, "\t")
		if err := module.CheckImportPath(pkg); err != nil {
			t.Errorf("Tool package %q %v", pkg, err)
		}
	}

	for _, tt := range []struct {
		prefix string
		want   []string
		desc   []string
	}{
		{".", nil, nil},
		{"./", nil, nil},
		{"gopls", []string{"golang.org/x/tools/gopls@"}, []string{"Go language server"}},
		{"golang.org/x/tools/g", []string{"golang.org/x/tools/gopls@"}, []string{"Go language server"}},
		{"github.com/golangci/", []string{"github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"}, []string{"Go linter collection"}},
	} {
		t.Run(tt.prefix, func(t *testing.T) {
			result := map[string]string{}
			CompleteToolPackages(t.Context(), result, tt.prefix)
			got := slices.Sorted(maps.Keys(result))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("CompleteToolPackages(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
			for i, pkg := range tt.want {
				if desc := result[pkg]; !strings.HasPrefix(desc, tt.desc[i]) {
					t.Fatalf("CompleteToolPackages(%q) %v = %v, want %v", tt.prefix, pkg, desc, tt.desc[i])
				}
			}
		})
	}
}

func TestAddTools(t *testing.T) {
	current := tools
	t.Cleanup(func() {
		tools = current
	})
	config := t.TempDir()
	if runtime.GOOS == "darwin" {
		t.Setenv("HOME", config)
		config += "/Library/Application Support"
		_ = os.MkdirAll(config, 0o700)
	}
	t.Setenv("XDG_CONFIG_HOME", config)

	if err := AddTools([]string{"example.com/tool/cmd/tool"}, []string{"Example tool"}); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(config, "golang-tool-completion", "tools.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "example.com/tool/cmd/tool\tExample tool\n") {
		t.Fatalf("tools file %q does not contain added tool: %q", path, string(data))
	}

	tools = ""
	result := map[string]string{}
	CompleteToolPackages(t.Context(), result, "example.com/tool/c")
	if result["example.com/tool/cmd/tool@"] != "Example tool" {
		t.Fatalf("CompleteToolPackages() = %#v, want added tool", result)
	}
}
