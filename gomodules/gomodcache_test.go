package gomodules

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCompleteModulesFromCache(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("GOMODCACHE", cache)
	mkdir := func(elem ...string) {
		t.Helper()
		must(t, os.MkdirAll(filepath.Join(append([]string{cache}, elem...)...), 0755))
	}
	write := func(name, data string) {
		t.Helper()
		must(t, os.WriteFile(filepath.Join(cache, filepath.FromSlash(name)), []byte(data), 0644))
	}

	mkdir("example.com")
	mkdir("example.com", "one@v1.0.0")
	mkdir("cache", "download", "example.com", "one", "@v")
	mkdir("cache", "download", "example.com", "one", "two", "@v")
	write("cache/download/example.com/one/@v/list", "v1.0.0\nv1.2.0\n")
	write("cache/download/example.com/one/@v/v1.3.0.info", "{}")
	write("cache/download/example.com/one/two/@v/list", "v0.1.0\n")

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
			got := keys(CompleteModules(tt.prefix))
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("CompleteModules(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestCompletePackagesFromCache(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("GOMODCACHE", cache)
	mkdir := func(elem ...string) {
		t.Helper()
		must(t, os.MkdirAll(filepath.Join(append([]string{cache}, elem...)...), 0755))
	}
	write := func(name, data string) {
		t.Helper()
		must(t, os.MkdirAll(filepath.Dir(filepath.Join(cache, filepath.FromSlash(name))), 0755))
		must(t, os.WriteFile(filepath.Join(cache, filepath.FromSlash(name)), []byte(data), 0644))
	}

	write("example.com/one@v1.2.0/one.go", "package one\n")
	write("example.com/one@v1.2.0/cmd/tool/main.go", "package main\n")
	write("example.com/one@v1.2.0/internal/hidden/hidden.go", "package hidden\n")
	write("example.com/one@v1.2.0/testdata/fixture/ignored.go", "package fixture\n")
	write("example.com/one/v2@v2.0.0/one.go", "package one\n")
	write("example.com/one/v2@v2.0.0/sub/sub.go", "package sub\n")
	write("github.com/!azure/azure-sdk-for-go@v1.0.0/sdk/resourcemanager/compute/compute.go", "package compute\n")
	mkdir("cache", "download", "example.com", "one", "@v")
	mkdir("cache", "download", "example.com", "one", "v2", "@v")
	write("cache/download/example.com/one/@v/list", "v1.2.0\nv1.3.0\n")
	write("cache/download/example.com/one/v2/@v/list", "v2.0.0\nv2.1.0\n")

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
			got := keys(CompletePackages(tt.prefix))
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("CompletePackages(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
		})
	}
}

func keys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
