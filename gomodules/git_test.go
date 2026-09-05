package gomodules

import (
	"testing"
)

func TestModulePathFromGitURL(t *testing.T) {
	tests := map[string]string{
		"https://github.com/example/project.git": "github.com/example/project",
		"ssh://git@gitlab.com/group/project.git": "gitlab.com/group/project",
		"git@github.com:example/project.git":     "github.com/example/project",
		"git://example.com/team/project":         "example.com/team/project",
	}
	for remote, want := range tests {
		t.Run(remote, func(t *testing.T) {
			got, err := modulePathFromGitURL(remote)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("modulePathFromGitURL(%q) = %q, want %q", remote, got, want)
			}
		})
	}
}
