package gomodules

import (
	"context"
	"errors"
	"net/url"
	"os/exec"
	"path"
	"strings"
)

// CompleteGitModule suggests a module path based on the current branch's
// upstream repository and the current directory's path within that repository.
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

func gitModulePath(ctx context.Context, dir string) (string, error) {
	branch, err := runGit(ctx, dir, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	remote, err := runGit(ctx, dir, "config", "--get", "branch."+branch+".remote")
	if err != nil {
		return "", err
	}
	if _, err := runGit(ctx, dir, "config", "--get", "branch."+branch+".merge"); err != nil {
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
	prefix, err := runGit(ctx, dir, "rev-parse", "--show-prefix")
	if err != nil {
		return "", err
	}
	return path.Join(repository, prefix), nil
}

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

func modulePathFromGitURL(remote string) (string, error) {
	// Git's scp-like syntax (git@example.com:owner/repo.git) is not understood
	// by net/url, so normalize it to an ssh URL first.
	if !strings.Contains(remote, "://") {
		if colon := strings.IndexByte(remote, ':'); colon > 0 && strings.Contains(remote[:colon], "@") {
			remote = "ssh://" + remote[:colon] + "/" + remote[colon+1:]
		}
	}
	u, err := url.Parse(remote)
	if err != nil || u.Hostname() == "" {
		if err == nil {
			err = &url.Error{Op: "parse", URL: remote, Err: errInvalidGitRemote}
		}
		return "", err
	}
	repository := strings.TrimSuffix(strings.Trim(u.Path, "/"), ".git")
	if repository == "" {
		return "", &url.Error{Op: "parse", URL: remote, Err: errInvalidGitRemote}
	}
	return path.Join(u.Hostname(), repository), nil
}

var errInvalidGitRemote = errors.New("git remote has no host or repository path")
