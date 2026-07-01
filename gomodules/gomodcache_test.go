package gomodules

import (
	"context"
	"embed"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

//go:embed testdata/module-cache
var embeddedTestdata embed.FS

func TestCompleteModulesFromCache(t *testing.T) {
	cache := testModCache(t, "testdata/module-cache")

	for _, tt := range []struct {
		prefix string
		want   []string
	}{
		{".", []string{}},
		{"./", []string{}},
		{"x", []string{}},
		{"", []string{"example.com/"}},
		{"e", []string{"example.com/"}},
		{"example.com", []string{"example.com/"}},
		{"example.com/", []string{"example.com/ABC/", "example.com/one/", "example.com/one@"}},
		{"example.com/ABC/", []string{"example.com/ABC/some/", "example.com/ABC/some@"}},
		{"example.com/o", []string{"example.com/one/", "example.com/one@"}},
		{"example.com/one/", []string{"example.com/one/two/", "example.com/one/two@", "example.com/one/v2/", "example.com/one/v2@"}},
		{"example.com/one@v", []string{"example.com/one@v1.0.0", "example.com/one@v1.1.0"}},
		{"example.com/one@l", []string{"example.com/one@latest"}},
		{"example.com/one@p", []string{"example.com/one@patch"}},
		{"example.com/one/t", []string{"example.com/one/two/", "example.com/one/two@"}},
		{"example.com/one/two@v", []string{"example.com/one/two@v1.0.0"}},
		{"example.com/one/v", []string{"example.com/one/v2/", "example.com/one/v2@"}},
		{"example.com/one/v2@v", []string{"example.com/one/v2@v2.0.0", "example.com/one/v2@v2.1.0"}},
	} {
		t.Run(tt.prefix, func(t *testing.T) {
			result := map[string]string{}
			cache.CompleteModules(result, tt.prefix)
			got := slices.Sorted(maps.Keys(result))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("CompleteModules(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestCompletePackagesFromCache(t *testing.T) {
	cache := testModCache(t, "testdata/module-cache")

	for _, tt := range []struct {
		prefix string
		want   []string
	}{
		{".", []string{}},
		{"./", []string{}},
		{"x", []string{}},
		{"", []string{"example.com/"}},
		{"e", []string{"example.com/"}},
		{"example.com", []string{"example.com/"}},
		{"example.com/", []string{"example.com/ABC/", "example.com/one/", "example.com/one@"}},
		{"example.com/o", []string{"example.com/one/", "example.com/one@"}},
		{"example.com/one/", []string{"example.com/one/pkg/", "example.com/one/two/", "example.com/one/two@", "example.com/one/v2/", "example.com/one/v2@"}},
		{"example.com/one@v", []string{"example.com/one@v1.0.0", "example.com/one@v1.1.0"}},
		{"example.com/one@l", []string{"example.com/one@latest"}},
		{"example.com/one@p", []string{"example.com/one@patch"}},
		{"example.com/one/v", []string{"example.com/one/v2/", "example.com/one/v2@"}},
		{"example.com/one/p", []string{"example.com/one/pkg/"}},
		{"example.com/one/t", []string{"example.com/one/two/", "example.com/one/two@"}},
		{"example.com/one/two/", []string{"example.com/one/two/pkg/"}},
		{"example.com/one/two@v", []string{"example.com/one/two@v1.0.0", "example.com/one/two@v1.1.0"}},
		{"example.com/A", []string{"example.com/ABC/"}},
		{"example.com/ABC/", []string{"example.com/ABC/some/", "example.com/ABC/some@"}},
	} {
		t.Run(tt.prefix, func(t *testing.T) {
			result := map[string]string{}
			cache.CompletePackages(result, tt.prefix)
			got := slices.Sorted(maps.Keys(result))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("CompletePackages(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
		})
	}
}

func testModCache(t *testing.T, dir string) ModCache {
	t.Helper()

	cache, err := fs.Sub(embeddedTestdata, dir)
	if err != nil {
		t.Fatal(err)
	}
	readDirFS, ok := cache.(fs.ReadDirFS)
	if !ok {
		t.Fatal("embedded module cache does not implement fs.ReadDirFS")
	}
	return ModCache{
		fs:  readDirFS,
		log: t.Logf,
	}
}

func TestCompleteProgramsFromNewestCachedModule(t *testing.T) {
	root := t.TempDir()
	writeFile := func(name, data string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(filepath.Join(root, "example.com", "prog@v1.0.0", "go.mod"), "module example.com/prog\n\ngo 1.23\n")
	writeFile(filepath.Join(root, "example.com", "prog@v1.0.0", "cmd", "old", "main.go"), "package main\nfunc main() {}\n")
	writeFile(filepath.Join(root, "example.com", "prog@v1.2.0", "go.mod"), "module example.com/prog\n\ngo 1.23\n")
	writeFile(filepath.Join(root, "example.com", "prog@v1.2.0", "main.go"), "package main\nfunc main() {}\n")
	writeFile(filepath.Join(root, "example.com", "prog@v1.2.0", "cmd", "new", "main.go"), "package main\nfunc main() {}\n")
	writeFile(filepath.Join(root, "example.com", "prog@v1.2.0", "pkg", "pkg.go"), "package pkg\n")

	cache := ModCache{fs: os.DirFS(root).(fs.ReadDirFS), root: root, log: t.Logf}
	result := map[string]string{}
	cache.CompletePrograms(context.Background(), result, "example.com/prog/c")

	got := slices.Sorted(maps.Keys(result))
	want := []string{"example.com/prog/cmd/", "example.com/prog/cmd/new"}
	if !slices.Equal(got, want) {
		t.Fatalf("CompletePrograms() = %v, want %v", got, want)
	}
	if result["example.com/prog/cmd/new"] != "main" {
		t.Fatalf("main package statuses = %#v", result)
	}
}
