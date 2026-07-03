package gomodules

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCompletePresentModules(t *testing.T) {
	root := writeTestModule(t)
	withChangeDirectory(t, root)

	result := map[string]string{}
	CompletePresentModules(context.Background(), result, "example.com/")

	got := slices.Sorted(maps.Keys(result))
	want := []string{"example.com/dep"}
	if !slices.Equal(got, want) {
		t.Fatalf("CompletePresentModules() = %v, want %v", got, want)
	}
	if result["example.com/dep"] != "module" {
		t.Fatalf("CompletePresentModules() statuses = %#v", result)
	}
}

func TestCompleteLocalMainPackages(t *testing.T) {
	root := writeTestModule(t)
	withChangeDirectory(t, root)

	result := map[string]string{}
	CompleteMainPackages(context.Background(), result, "./")

	got := slices.Sorted(maps.Keys(result))
	want := []string{"./cmd/tool"}
	if !slices.Equal(got, want) {
		t.Fatalf("CompleteMainPackages() = %v, want %v", got, want)
	}
	if result["./cmd/tool"] != "local main" {
		t.Fatalf("CompleteMainPackages() statuses = %#v", result)
	}
}

func writeTestModule(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), `module example.com/root

go 1.25

require example.com/dep v0.0.0

replace example.com/dep => ./dep
`)
	writeFile(t, filepath.Join(root, "cmd", "tool", "main.go"), "package main\nfunc main() {}\n")
	writeFile(t, filepath.Join(root, "pkg", "pkg.go"), "package pkg\n")
	writeFile(t, filepath.Join(root, "dep", "go.mod"), "module example.com/dep\n\ngo 1.25\n")
	writeFile(t, filepath.Join(root, "dep", "dep.go"), "package dep\n")
	return root
}

func writeFile(t *testing.T, name string, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func withChangeDirectory(t *testing.T, dir string) {
	t.Helper()
	old := ChangeDirectory
	ChangeDirectory = dir
	t.Cleanup(func() { ChangeDirectory = old })
}
