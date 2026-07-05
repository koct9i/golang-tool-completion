package completion

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestCompletionInstallCreatesParentDirectories(t *testing.T) {
	dataHome := filepath.Join(t.TempDir(), "xdg-data")
	t.Setenv("XDG_DATA_HOME", dataHome)

	var output bytes.Buffer
	if err := doCompletionScript(&output, "bash", "go", "golang-tool-completion", true); err != nil {
		t.Fatal(err)
	}

	scriptPath := filepath.Join(dataHome, "bash-completion", "completions", "go")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "completion --complete bash") {
		t.Fatalf("installed script %q does not look like bash completion: %q", scriptPath, data)
	}
	if !strings.Contains(output.String(), scriptPath) {
		t.Fatalf("install output %q does not mention %q", output.String(), scriptPath)
	}
}

func TestCompletionStateDoesNotLeakBetweenRequests(t *testing.T) {
	stopOnFirstArg := 1
	var output bytes.Buffer
	root := &cli.Command{
		Name:   "go",
		Writer: &output,
		Commands: []*cli.Command{
			{
				Name:         "run",
				Flags:        []cli.Flag{&cli.BoolFlag{Name: "v", Usage: "verbose"}},
				StopOnNthArg: &stopOnFirstArg,
				Arguments: []cli.Argument{
					&Argument{
						Name: "package",
						Max:  -1,
						OnComplete: func(ctx context.Context, result map[string]string, prefix string) {
							result[prefix+"/"] = "package"
						},
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					LastCommand = c
					return nil
				},
			},
		},
	}

	if err := doCompletion(context.Background(), root, "bash", []string{"run", "example"}); err != nil {
		t.Fatal(err)
	}
	output.Reset()

	if err := doCompletion(context.Background(), root, "bash", []string{"run", "-"}); err != nil {
		t.Fatal(err)
	}

	if got := output.String(); !strings.Contains(got, "-v") {
		t.Fatalf("second completion = %q, want flag completion; parsed args leaked: %#v", got, ParsedArguments)
	}
}

func TestCompletionCommandEndToEnd(t *testing.T) {
	moduleDir, err := filepath.Abs(filepath.Join("testdata", "e2e-module"))
	if err != nil {
		t.Fatal(err)
	}
	handler := buildCompletionHandler(t)
	modCacheDir, err := filepath.Abs(filepath.Join("testdata", "module-cache"))
	if err != nil {
		t.Fatal(err)
	}
	configDir := writeCompletionConfig(t, map[string]string{
		"trending.txt": "example.com/substring-module\tSubstring module\n",
		"tools.txt":    "example.com/acme/cmd/subtool\tSubstring tool\n",
	})

	tests := []struct {
		name          string
		args          []string
		gomodcache    string
		xdgConfigHome string
		want          []string
		notWant       []string
	}{
		{
			name: "root command prefix",
			args: []string{"bu"},
			want: []string{"build"},
		},
		{
			name: "command flag prefix",
			args: []string{"build", "-ra"},
			want: []string{"-race"},
		},
		{
			name:          "tool package substring",
			args:          []string{"install", "subtool"},
			xdgConfigHome: configDir,
			want:          []string{"example.com/acme/cmd/subtool@"},
		},
		{
			name:          "trending module substring",
			args:          []string{"mod", "download", "substring"},
			xdgConfigHome: configDir,
			want:          []string{"example.com/substring-module@"},
		},
		{
			name: "used module substring",
			args: []string{"-C", moduleDir, "get", "-u", "dep"},
			want: []string{"example.com/e2edep@"},
		},
		{
			name:    "short substring does not match used module",
			args:    []string{"-C", moduleDir, "get", "-u", "ep"},
			notWant: []string{"example.com/e2edep@"},
		},
		{
			name: "known tool package",
			args: []string{"install", "gopls"},
			want: []string{"golang.org/x/tools/gopls@"},
		},
		{
			name:       "cached main package",
			args:       []string{"install", "example.com/e2emod/c"},
			gomodcache: modCacheDir,
			want:       []string{"example.com/e2emod/cmd/app@"},
		},
		{
			name:       "package becomes module",
			args:       []string{"clean", "example.com/split/pkg@v"},
			gomodcache: modCacheDir,
			want: []string{
				"example.com/split/pkg@v1.0.0",
				"example.com/split/pkg@v2.0.0",
			},
		},
		{
			name:       "package appears in later module version",
			args:       []string{"clean", "example.com/later/newpkg@v"},
			gomodcache: modCacheDir,
			want:       []string{"example.com/later/newpkg@v1.1.0"},
			notWant:    []string{"example.com/later/newpkg@v1.0.0"},
		},
		{
			name:       "cached module",
			args:       []string{"mod", "download", "example.com/e2em"},
			gomodcache: modCacheDir,
			want:       []string{"example.com/e2emod@"},
		},
		{
			name:       "cached package",
			args:       []string{"clean", "example.com/e2emod/p"},
			gomodcache: modCacheDir,
			want:       []string{"example.com/e2emod/pkg/"},
		},
		{
			name: "local main package",
			args: []string{"-C", moduleDir, "run", "./"},
			want: []string{"./cmd/tool"},
		},
		{
			name: "build dependency package",
			args: []string{"-C", moduleDir, "build", "example.com/e2ed"},
			want: []string{"example.com/e2edep"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runCompletionCommand(t, handler, tt.gomodcache, tt.xdgConfigHome, tt.args...)
			for _, want := range tt.want {
				if !completionOutputContains(got, want) {
					t.Fatalf("completion output for args %q = %q, want %q", tt.args, got, want)
				}
			}
			for _, notWant := range tt.notWant {
				if completionOutputContains(got, notWant) {
					t.Fatalf("completion output for args %q = %q, did not want %q", tt.args, got, notWant)
				}
			}
		})
	}
}

func writeCompletionConfig(t *testing.T, files map[string]string) string {
	t.Helper()

	configDir := filepath.Join(t.TempDir(), "config")
	configCompletionDir := filepath.Join(configDir, "golang-tool-completion")
	if err := os.MkdirAll(configCompletionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(configCompletionDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return configDir
}

func buildCompletionHandler(t *testing.T) string {
	t.Helper()

	handler := filepath.Join(t.TempDir(), "golang-tool-completion-e2e")
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", handler, "..")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	cmd.Dir = "."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build -o %s .. failed: %v\nstdout:\n%s\nstderr:\n%s", handler, err, stdout.String(), stderr.String())
	}
	return handler
}

func runCompletionCommand(t *testing.T, handler string, gomodcache string, xdgConfigHome string, args ...string) string {
	t.Helper()

	cmdArgs := append([]string{"completion", "--complete", "bash", "--"}, args...)
	cmd := exec.CommandContext(t.Context(), handler, cmdArgs...)
	if xdgConfigHome == "" {
		xdgConfigHome = t.TempDir()
	}
	env := append(os.Environ(),
		"HOME="+t.TempDir(),
		"XDG_CONFIG_HOME="+xdgConfigHome,
		"GOWORK=off",
	)
	if gomodcache != "" {
		env = append(env, "GOMODCACHE="+gomodcache)
	}
	cmd.Env = env
	cmd.Dir = "."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %s failed: %v\nstdout:\n%s\nstderr:\n%s", handler, strings.Join(cmdArgs, " "), err, stdout.String(), stderr.String())
	}
	return stdout.String()
}

func completionOutputContains(output, want string) bool {
	for line := range strings.Lines(output) {
		candidate := strings.TrimSpace(line)
		if before, _, found := strings.Cut(candidate, " ("); found {
			candidate = strings.TrimSpace(before)
		}
		if candidate == want {
			return true
		}
	}
	return false
}
