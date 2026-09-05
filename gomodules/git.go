package gomodules

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"strings"
)

func CompleteGitModule(ctx context.Context, result map[string]string, prefix string) {
	dir := "."
	if ChangeDirectory != "" {
		dir = expandPath(ChangeDirectory)
	}
	module, err := gitModulePath(ctx, dir)
	if err == nil && strings.HasPrefix(module, prefix) {
		result[module] = "git upstream"
	}
}

func modulePathFromGitURL(remote string) (string, error) {
	// git@example.com:owner/repo.git -> ssh://example.com/owner/repo.git
	if !strings.Contains(remote, "://") {
		if host, path, ok := strings.Cut(remote, ":"); ok {
			if _, h, ok := strings.Cut(host, "@"); ok {
				host = h
			}
			remote = "ssh://" + host + "/" + path
		}
	}
	u, err := url.Parse(remote)
	if err != nil {
		return "", err
	}
	hostname := u.Hostname()
	repository := strings.TrimSuffix(strings.Trim(u.Path, "/"), ".git")
	if hostname == "" || repository == "" {
		return "", fmt.Errorf("git remote has no host or repository")
	}
	return path.Join(hostname, repository), nil
}

func gitModulePath(ctx context.Context, dir string) (string, error) {
	prefix, err := runGit(ctx, dir, "rev-parse", "--show-prefix")
	if err != nil {
		return "", err
	}
	branch, err := runGit(ctx, dir, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	remote, err := runGit(ctx, dir, "config", "--get", "branch."+branch+".remote")
	if err != nil {
		return "", err
	}
	remoteURL, err := runGit(ctx, dir, "remote", "get-url", remote)
	if err != nil {
		return "", err
	}
	repository, err := modulePathFromGitURL(remoteURL)
	if err != nil {
		return "", err
	}
	return path.Join(repository, prefix), nil
}

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	//nolint:gosec //command
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}
