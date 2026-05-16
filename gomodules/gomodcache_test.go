package gomodules

import (
	"embed"
	"io/fs"
	"slices"
	"testing"
)

//go:embed testdata/module-cache testdata/package-cache
var embeddedTestdata embed.FS

func TestCompleteModulesFromCache(t *testing.T) {
	cache := testModCache(t, "testdata/module-cache")

	for _, tt := range []struct {
		prefix string
		want   []string
	}{
		{"example.c", []string{"example.com/"}},
		{"example.com/o", []string{"example.com/one@"}},
		{"example.com/one/t", []string{"example.com/one/two@"}},
		{"example.com/one@v1.", []string{"example.com/one@v1.0.0", "example.com/one@v1.2.0", "example.com/one@v1.3.0"}},
		{"example.com/one@lat", []string{"example.com/one@latest"}},
	} {
		t.Run(tt.prefix, func(t *testing.T) {
			got := keys(cache.CompleteModules(tt.prefix))
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("CompleteModules(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestCompletePackagesFromCache(t *testing.T) {
	cache := testModCache(t, "testdata/package-cache")

	for _, tt := range []struct {
		prefix string
		want   []string
	}{
		{"example.com/one/cmd/t", []string{"example.com/one/cmd/tool"}},
		{"example.com/one/v", []string{"example.com/one/v2", "example.com/one/v2/", "example.com/one/v2/sub"}},
		{"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/com", []string{"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute"}},
		{"example.com/one/cmd/tool@v1.", []string{"example.com/one/cmd/tool@v1.2.0", "example.com/one/cmd/tool@v1.3.0"}},
		{"example.com/one/v2/sub@v2.", []string{"example.com/one/v2/sub@v2.0.0", "example.com/one/v2/sub@v2.1.0"}},
	} {
		t.Run(tt.prefix, func(t *testing.T) {
			got := keys(cache.CompletePackages(tt.prefix))
			slices.Sort(got)
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
	return ModCache{fs: readDirFS}
}

func keys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
