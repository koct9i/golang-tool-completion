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
		{"golang.or", []string{"golang.org/"}, []string{"trending"}},
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
				if desc := result[mod]; desc != tt.desc[i] {
					t.Fatalf("CompleteTrending(%q) %v = %v, want %v", tt.prefix, mod, desc, tt.desc[i])
				}
			}
		})
	}
}
