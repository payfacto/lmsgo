package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const askSystem = "You are a precise code analyst. " +
	"Read the provided source files carefully and answer the question concisely. " +
	"Cite file names and line numbers where relevant. " +
	"Be direct — no preamble, no filler."

func runAsk(args []string) {
	fs_ := flag.NewFlagSet("ask", flag.ExitOnError)
	question := fs_.String("question", "", "Question to answer about the files (required)")
	glob := fs_.String("glob", "", "Glob pattern when a directory is supplied")
	maxTokens := fs_.Int("max-tokens", 8192, "Max output tokens")

	fs_.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: lmsgo ask [flags] <file|dir> [<file|dir>...]")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs_.PrintDefaults()
	}

	if err := fs_.Parse(args); err != nil {
		os.Exit(1)
	}
	if *question == "" {
		fmt.Fprintln(os.Stderr, "error: --question is required")
		fs_.Usage()
		os.Exit(1)
	}
	paths := fs_.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "error: at least one path is required")
		fs_.Usage()
		os.Exit(1)
	}

	files, err := collectFiles(paths, *glob)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "error: no files found")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "[lmsgo ask] reading %d file(s)…\n", len(files))
	corpus := buildCorpus(files)

	answer, err := complete([]message{
		{Role: "system", Content: askSystem},
		{Role: "user", Content: "<corpus>\n" + corpus + "\n</corpus>"},
		{Role: "user", Content: *question},
	}, *maxTokens, 0.1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(answer)
}

func collectFiles(paths []string, pattern string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[lmsgo] warning: %s not found, skipping\n", p)
			continue
		}
		if !info.IsDir() {
			files = append(files, p)
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			if pattern != "" {
				matched, _ := filepath.Match(pattern, filepath.Base(path))
				if !matched {
					return nil
				}
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", p, err)
		}
	}
	return files, nil
}

func buildCorpus(files []string) string {
	var sb strings.Builder
	for i, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[lmsgo] warning: could not read %s: %v\n", f, err)
			continue
		}
		if i > 0 {
			sb.WriteString("\n\n")
		}
		fmt.Fprintf(&sb, "<file path='%s'>\n%s\n</file>", f, content)
	}
	return sb.String()
}
