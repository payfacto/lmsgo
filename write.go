package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const writeSystem = "You are an expert software engineer. " +
	"Generate the requested file based on the spec and any provided context. " +
	"Output ONLY the file contents — no markdown fences, no explanation, no preamble. " +
	"Match the style and conventions of the context files exactly."

func runWrite(args []string) {
	fs_ := flag.NewFlagSet("write", flag.ExitOnError)
	spec := fs_.String("spec", "", "Description of the file to generate (required)")
	target := fs_.String("target", "", "Output file path to write (required)")
	maxTokens := fs_.Int("max-tokens", 16384, "Max output tokens")
	dryRun := fs_.Bool("dry-run", false, "Print generated content instead of writing to disk")

	fs_.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: lmsgo write [flags] [context-file...]")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs_.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nContext files are passed as positional arguments after flags.")
	}

	if err := fs_.Parse(args); err != nil {
		os.Exit(1)
	}
	if *spec == "" {
		fmt.Fprintln(os.Stderr, "error: --spec is required")
		fs_.Usage()
		os.Exit(1)
	}
	if *target == "" {
		fmt.Fprintln(os.Stderr, "error: --target is required")
		fs_.Usage()
		os.Exit(1)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Generate: %s\n", *spec)

	for _, ctxPath := range fs_.Args() {
		content, err := os.ReadFile(ctxPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[lmsgo] warning: context file %s not found\n", ctxPath)
			continue
		}
		fmt.Fprintf(&sb, "\n<context file='%s'>\n%s\n</context>", ctxPath, content)
	}

	fmt.Fprintf(&sb, "\n\nOutput file: %s", *target)

	fmt.Fprintf(os.Stderr, "[lmsgo write] generating %s…\n", *target)

	generated, err := complete([]message{
		{Role: "system", Content: writeSystem},
		{Role: "user", Content: sb.String()},
	}, *maxTokens, 0.2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *dryRun {
		fmt.Println(generated)
		return
	}

	if err := os.MkdirAll(filepath.Dir(*target), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: create directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*target, []byte(generated), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: write file: %v\n", err)
		os.Exit(1)
	}
	lines := strings.Count(generated, "\n") + 1
	fmt.Fprintf(os.Stderr, "[lmsgo write] wrote %d lines to %s\n", lines, *target)
}
