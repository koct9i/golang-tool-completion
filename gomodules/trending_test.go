package gomodules

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"golang.org/x/mod/module"
)

func TestTrendingModules(t *testing.T) {
	for line := range strings.Lines(trending) {
		modulePath, _, ok := parseTrendingLine(line)
		if !ok {
			continue
		}
		if err := module.CheckPath(modulePath); err != nil {
			t.Errorf("Trending %q %v", modulePath, err)
		}
	}

	for _, tt := range []struct {
		prefix string
		want   []string
	}{
		{".", []string{}},
		{"./", []string{}},
		{"example.com", []string{}},
		{"golang.or", []string{"golang.org/"}},
	} {
		t.Run(tt.prefix, func(t *testing.T) {
			result := map[string]string{}
			CompleteTrending(result, tt.prefix)
			got := slices.Sorted(maps.Keys(result))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("CompleteTrending(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
			if len(result) > 0 {
				status := slices.Compact(slices.Collect(maps.Values(result)))
				if len(status) != 1 || status[0] != "trending" {
					t.Fatalf("CompleteTrending status %v", status)
				}
			}
		})
	}
}

func TestCompleteTrendingDescription(t *testing.T) {
	current := trending
	t.Cleanup(func() {
		trending = current
	})
	trending = "example.com/mod\tExample module\n"

	result := map[string]string{}
	CompleteTrending(result, "example.com/mod")

	if got := result["example.com/mod@"]; got != "Example module" {
		t.Fatalf("CompleteTrending description = %q, want %q", got, "Example module")
	}
}
