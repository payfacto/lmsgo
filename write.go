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
	"Your VERY FIRST CHARACTER must be the first character of the file itself. " +
	"Do NOT write any planning, reasoning, thinking aloud, plans, summaries, the filename, or commentary — not before, during, or after the file. " +
	"No markdown fences. No explanation. Each line of the file appears exactly once. " +
	"Stop immediately after the final line of the file. " +
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

	cleaned, didStrip := stripPreamble(generated)
	if didStrip {
		fmt.Fprintln(os.Stderr, "[lmsgo write] stripped model preamble before first code line")
	}
	generated = cleaned

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

// codeStartMarkers are the prefixes that mark the beginning of file contents
// for the languages and formats lmsgo write commonly produces. The list is
// scanned in stripPreamble to find where a model's planning preamble ends and
// the actual file begins.
var codeStartMarkers = []string{
	"package ", "import ", "from ", "func ", "class ", "def ",
	"function ", "const ", "let ", "var ", "using ", "namespace ",
	"#!", "<?", "<!DOCTYPE", "<html", "FROM ", "/*", "//",
}

// stripPreamble cleans up small-model output. Returns the cleaned string and
// a bool indicating whether a real preamble or fence was removed (trailing
// whitespace normalization alone does NOT count as stripping).
//  1. Removes any planning/reasoning text emitted before the file contents by
//     scanning for the first plausible code-start marker (e.g. "package ",
//     "import ", "#!", "<?").
//  2. Truncates at any markdown fence (```), since the prompt forbids fences
//     and they typically wrap a duplicated copy of the file.
func stripPreamble(s string) (string, bool) {
	bestIdx := -1
	for _, kw := range codeStartMarkers {
		idx := strings.Index(s, kw)
		if idx < 0 {
			continue
		}
		if idx > 0 && isWordByte(s[idx-1]) {
			continue
		}
		if bestIdx == -1 || idx < bestIdx {
			bestIdx = idx
		}
	}
	stripped := false
	if bestIdx > 0 {
		s = s[bestIdx:]
		stripped = true
	}
	if i := strings.Index(s, "\n```"); i >= 0 {
		s = s[:i]
		stripped = true
	}
	return strings.TrimRight(s, "\n") + "\n", stripped
}

// isWordByte reports whether b is an ASCII letter, digit, or underscore — i.e.
// a character that, when adjacent to a keyword, means the keyword is part of a
// longer identifier rather than a standalone token.
func isWordByte(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '_'
}
