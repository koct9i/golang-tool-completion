package completion

import (
	"bytes"
	"context"
	"os"
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
