package gomodules

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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

func TestGitModulePathIncludesDirectoryWithinRepository(t *testing.T) {
	repository := t.TempDir()
	runGitTestCommand(t, repository, "init", "--initial-branch=feature")
	runGitTestCommand(t, repository, "remote", "add", "upstream", "git@github.com:example/project.git")
	runGitTestCommand(t, repository, "config", "branch.feature.remote", "upstream")
	runGitTestCommand(t, repository, "config", "branch.feature.merge", "refs/heads/feature")

	dir := filepath.Join(repository, "tools", "generator")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := gitModulePath(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := "github.com/example/project/tools/generator"; got != want {
		t.Fatalf("gitModulePath() = %q, want %q", got, want)
	}
}

func TestGitModulePathRequiresBranchUpstream(t *testing.T) {
	repository := t.TempDir()
	runGitTestCommand(t, repository, "init", "--initial-branch=feature")
	runGitTestCommand(t, repository, "remote", "add", "origin", "https://github.com/example/project.git")
	runGitTestCommand(t, repository, "config", "branch.feature.remote", "origin")

	if _, err := gitModulePath(context.Background(), repository); err == nil {
		t.Fatal("gitModulePath() succeeded without an upstream merge ref")
	}
}

func runGitTestCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
