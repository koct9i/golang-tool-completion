package completion

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func requireExitCodeZero(t *testing.T, err error) {
	t.Helper()
	var exitCoder interface{ ExitCode() int }
	if !errors.As(err, &exitCoder) || exitCoder.ExitCode() != 0 {
		t.Fatalf("expected cli exit code 0, got %v", err)
	}
}

func TestDoCompletion_Commands(t *testing.T) {
	var out bytes.Buffer
	cmd := &cli.Command{
		Name:   "tool",
		Writer: &out,
		Commands: []*cli.Command{
			{Name: "build", Usage: "build project", Aliases: []string{"b"}},
			{Name: "bench", Usage: "run benchmark"},
			{Name: "secret", Hidden: true},
		},
	}

	err := doCompletion(context.Background(), cmd, "bash", nil)
	requireExitCodeZero(t, err)

	got := out.String()
	if !strings.Contains(got, "build") || !strings.Contains(got, "bench") || strings.Contains(got, "secret") {
		t.Fatalf("unexpected completion output: %q", got)
	}
}

func TestDoCompletion_FlagsBashFormatting(t *testing.T) {
	var out bytes.Buffer
	cmd := &cli.Command{
		Name:   "tool",
		Writer: &out,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Usage: "verbose output"},
			&cli.BoolFlag{Name: "version", Usage: "print version"},
			&cli.BoolFlag{Name: "v", Usage: "short"},
		},
	}

	err := doCompletion(context.Background(), cmd, "bash", []string{"--v"})
	requireExitCodeZero(t, err)

	got := out.String()
	if !strings.Contains(got, "--verbose") || !strings.Contains(got, "(verbose output)") {
		t.Fatalf("missing verbose suggestion: %q", got)
	}
	if !strings.Contains(got, "--version") || !strings.Contains(got, "(print version)") {
		t.Fatalf("missing version suggestion: %q", got)
	}
	if strings.Contains(got, "\n-v") {
		t.Fatalf("unexpected short flag for long prefix: %q", got)
	}
}

func TestDoCompletion_FishFormatting(t *testing.T) {
	var out bytes.Buffer
	cmd := &cli.Command{
		Name:   "tool",
		Writer: &out,
		Commands: []*cli.Command{{Name: "build", Usage: "build project"}},
	}

	err := doCompletion(context.Background(), cmd, "fish", nil)
	requireExitCodeZero(t, err)

	if out.String() != "build\tbuild project\n" {
		t.Fatalf("unexpected fish output: %q", out.String())
	}
}
