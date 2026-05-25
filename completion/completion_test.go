package completion

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestDoCompletionCommands(t *testing.T) {
	root := testRootCommand()
	root.Commands = []*cli.Command{
		{Name: "build", Usage: "compile packages"},
		{Name: "hidden", Hidden: true, Usage: "should not appear"},
	}

	got := runCompletion(t, root, "", []string{"b"})
	if got != "build\n" {
		t.Fatalf("doCompletion command suggestions = %q, want %q", got, "build\n")
	}
}

func TestDoCompletionCommandAliases(t *testing.T) {
	root := testRootCommand()
	root.Commands = []*cli.Command{
		{Name: "build", Aliases: []string{"bld"}, Usage: "compile packages"},
	}

	got := runCompletion(t, root, "", []string{"bld"})
	if got != "bld\n" {
		t.Fatalf("doCompletion alias suggestions = %q, want %q", got, "bld\n")
	}
}

func TestDoCompletionFlags(t *testing.T) {
	root := testRootCommand()
	root.Commands = []*cli.Command{
		{
			Name:   "build",
			Action: captureLastCommandAction,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "v", Usage: "verbose"},
				&cli.BoolFlag{Name: "verbose", Usage: "verbose output"},
			},
		},
	}

	got := runCompletion(t, root, "", []string{"build", "-"})
	want := "--verbose\n-v\n"
	if got != want {
		t.Fatalf("doCompletion flag suggestions = %q, want %q", got, want)
	}
}

func TestDoCompletionArguments(t *testing.T) {
	root := testRootCommand()
	root.Commands = []*cli.Command{
		{
			Name:   "get",
			Action: captureLastCommandAction,
			Arguments: []cli.Argument{
				&Argument{
					Name: "pkg",
					Max:  1,
					OnComplete: func(_ context.Context, prefix string) map[string]string {
						return map[string]string{
							"example.com/other": "other",
							"example.com/pkg":   "pkg",
						}
					},
				},
			},
		},
	}

	got := runCompletion(t, root, "fish", []string{"get", "exa"})
	want := "example.com/other\tother\nexample.com/pkg\tpkg\n"
	if got != want {
		t.Fatalf("doCompletion argument suggestions = %q, want %q", got, want)
	}
}

func TestDoCompletionNoCompletionAfterDoubleDash(t *testing.T) {
	root := testRootCommand()
	root.Commands = []*cli.Command{
		{Name: "build", Usage: "compile packages"},
	}

	got := runCompletion(t, root, "", []string{"build", "--", "x"})
	if got != "" {
		t.Fatalf("doCompletion suggestions after -- = %q, want empty", got)
	}
}

func testRootCommand() *cli.Command {
	return &cli.Command{
		Name:   "go",
		Writer: &bytes.Buffer{},
	}
}

func captureLastCommandAction(_ context.Context, c *cli.Command) error {
	if WithinCompletion {
		LastCommand = c
	}
	return nil
}

func runCompletion(t *testing.T, root *cli.Command, shell string, completeArgs []string) string {
	t.Helper()

	resetCompletionState()
	t.Cleanup(resetCompletionState)

	buffer, ok := root.Writer.(*bytes.Buffer)
	if !ok {
		t.Fatal("root writer must be *bytes.Buffer")
	}
	buffer.Reset()

	err := doCompletion(context.Background(), root, shell, completeArgs)
	if err == nil {
		t.Fatal("doCompletion returned nil, want cli.Exit error")
	}
	var exitCoder cli.ExitCoder
	if !errors.As(err, &exitCoder) || exitCoder.ExitCode() != 0 {
		t.Fatalf("doCompletion error = %v, want cli.Exit code 0", err)
	}

	return buffer.String()
}

func resetCompletionState() {
	WithinCompletion = false
	LastCommand = nil
	ParsedArguments = nil
	ArgumentCompletors = nil
}
