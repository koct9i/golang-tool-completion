package gomodules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompleteDocPackagesIncludesStandardLibrary(t *testing.T) {
	result := CompleteDocPackages("fm")
	if _, ok := result["fmt"]; !ok {
		t.Fatalf("expected std package fmt in results, got %v", result)
	}
}

func TestCompleteDocPackagesIncludesLocalPackages(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, "cmd"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmp, "internal"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmp, "_hidden"), 0755); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	result := CompleteDocPackages("./")
	if _, ok := result["./cmd/"]; !ok {
		t.Fatalf("expected local package ./cmd/ in results, got %v", result)
	}
	if _, ok := result["./internal/"]; !ok {
		t.Fatalf("expected local package ./internal/ in results, got %v", result)
	}
	if _, ok := result["./_hidden/"]; ok {
		t.Fatalf("did not expect hidden package ./_hidden/ in results, got %v", result)
	}
}

func TestCompleteDocPackagesIncludesPackageSymbols(t *testing.T) {
	result := CompleteDocPackages("fmt.Pr")
	if _, ok := result["fmt.Printf"]; !ok {
		t.Fatalf("expected symbol fmt.Printf in results, got %v", result)
	}
}

func TestSplitDocSymbolPrefix(t *testing.T) {
	pkg, sym, ok := splitDocSymbolPrefix("example.com/pkg.Type.M")
	if !ok || pkg != "example.com/pkg" || sym != "Type.M" {
		t.Fatalf("unexpected split: pkg=%q sym=%q ok=%v", pkg, sym, ok)
	}
}
