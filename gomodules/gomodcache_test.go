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
	root := absTestdataPath(t, "program-cache")

	//nolint:errcheck //ok
	cache := ModCache{fs: os.DirFS(root).(fs.ReadDirFS), dir: root, log: t.Logf}
	result := map[string]string{}
	cache.CompleteMainPackages(context.Background(), result, "example.com/prog/c")

	got := slices.Sorted(maps.Keys(result))
	want := []string{"example.com/prog/cmd/new@"}
	if !slices.Equal(got, want) {
		t.Fatalf("CompletePrograms() = %v, want %v", got, want)
	}
	if result["example.com/prog/cmd/new@"] != "cache" {
		t.Fatalf("main package statuses = %#v", result)
	}
}

func TestCompletePresentModules(t *testing.T) {
	root := absTestdataPath(t, "local-module")
	withChangeDirectory(t, root)

	result := map[string]string{}
	CompleteUsedModules(context.Background(), result, "example.com/")

	got := slices.Sorted(maps.Keys(result))
	want := []string{"example.com/dep", "example.com/dep@"}
	if !slices.Equal(got, want) {
		t.Fatalf("CompletePresentModules() = %v, want %v", got, want)
	}
	if result["example.com/dep"] != "used" {
		t.Fatalf("CompletePresentModules() statuses = %#v", result)
	}
}

func TestCompleteLocalMainPackages(t *testing.T) {
	root := absTestdataPath(t, "local-module")
	withChangeDirectory(t, root)

	result := map[string]string{}
	CompleteMainPackages(context.Background(), result, "./")

	got := slices.Sorted(maps.Keys(result))
	want := []string{"./cmd/tool"}
	if !slices.Equal(got, want) {
		t.Fatalf("CompleteMainPackages() = %v, want %v", got, want)
	}
	if result["./cmd/tool"] != "tool" {
		t.Fatalf("CompleteMainPackages() statuses = %#v", result)
	}
}

func TestCompleteUsedModuleVersions(t *testing.T) {
	root := absTestdataPath(t, "local-module")
	withChangeDirectory(t, root)

	result := map[string]string{}
	CompleteUsedModules(context.Background(), result, "example.com/dep@v")

	got := slices.Sorted(maps.Keys(result))
	want := []string{"example.com/dep@v0.0.0"}
	if !slices.Equal(got, want) {
		t.Fatalf("CompleteUsedModules() = %v, want %v", got, want)
	}
	if result["example.com/dep@v0.0.0"] != "used" {
		t.Fatalf("CompleteUsedModules() statuses = %#v", result)
	}
}

func TestCompleteDependenciesFromLocalModule(t *testing.T) {
	root := absTestdataPath(t, "local-module")
	withChangeDirectory(t, filepath.Join(root, "cmd", "tool"))

	result := map[string]string{}
	CompleteDependencies(context.Background(), result, "example.com/root/c")

	got := slices.Sorted(maps.Keys(result))
	want := []string{"example.com/root/cmd/tool"}
	if !slices.Equal(got, want) {
		t.Fatalf("CompleteDependencies() = %v, want %v", got, want)
	}
	if result["example.com/root/cmd/tool"] != "deps" {
		t.Fatalf("CompleteDependencies() statuses = %#v", result)
	}
}

func absTestdataPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func withChangeDirectory(t *testing.T, dir string) {
	t.Helper()
	old := ChangeDirectory
	ChangeDirectory = dir
	t.Cleanup(func() { ChangeDirectory = old })
}
