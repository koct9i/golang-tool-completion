package gomodules

import (
	"embed"
	"io/fs"
	"maps"
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

func TestCompleteGoVersions(t *testing.T) {
	cache := testModCache(t, "testdata/module-cache")

	for _, tt := range []struct {
		prefix string
		want   []string
	}{
		{"go@", []string{"go@1.22.0", "go@1.22.1", "go@1.23.0", "go@1.23rc1", "go@latest", "go@patch"}},
		{"go@1.22", []string{"go@1.22.0", "go@1.22.1"}},
		{"go@1.23", []string{"go@1.23.0", "go@1.23rc1"}},
		{"go@1.23r", []string{"go@1.23rc1"}},
		{"go@l", []string{"go@latest"}},
		{"go@p", []string{"go@patch"}},
		{"toolchain@", []string{"toolchain@1.22.0", "toolchain@1.22.1", "toolchain@1.23.0", "toolchain@1.23rc1", "toolchain@latest", "toolchain@patch"}},
		{"toolchain@1.22", []string{"toolchain@1.22.0", "toolchain@1.22.1"}},
		{"toolchain@1.23", []string{"toolchain@1.23.0", "toolchain@1.23rc1"}},
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

func TestExtractGoVersion(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  string
		ok    bool
	}{
		{"v0.0.1-go1.22.0.linux-amd64", "1.22.0", true},
		{"v0.0.1-go1.22.1.darwin-arm64", "1.22.1", true},
		{"v0.0.1-go1.23rc1.linux-amd64", "1.23rc1", true},
		{"v0.0.1-go1.23.0.windows-amd64", "1.23.0", true},
		{"v1.2.3", "", false},
		{"v0.0.1-go.linux-amd64", "", false},
	} {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := extractGoVersion(tt.input)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("extractGoVersion(%q) = (%q, %v), want (%q, %v)", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}
