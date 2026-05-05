package main

import (
	"fmt"
	"os"

	"github.com/payfacto/lmsgo/internal/version"
)

const usage = `lmsgo — delegate bulk I/O from Claude Code to a local LM Studio model

Usage:
  lmsgo ask    --question "..." <file|dir> [<file|dir>...]
  lmsgo write  --spec "..." --target <output> [context-file...]
  lmsgo extract [-o output] <session.jsonl>
  lmsgo setup  [--model NAME] [--dry-run]
  lmsgo version

Subcommands:
  ask       Read files and answer a question using the local model
  write     Generate a boilerplate file from a spec
  extract   Strip a Claude Code JSONL session to readable text
  setup     Detect models, write env file, patch ~/.claude/CLAUDE.md
  version   Print the version and exit

Environment:
  LMS_BASE_URL   LM Studio API base URL  (default: http://localhost:1234/v1)
  LMS_MODEL      Model ID                (default: local-model)
  LMS_API_KEY    API key                 (default: lm-studio — ignored by LM Studio)

Run 'lmsgo <subcommand> --help' for subcommand flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "ask":
		runAsk(os.Args[2:])
	case "write":
		runWrite(os.Args[2:])
	case "extract":
		runExtract(os.Args[2:])
	case "setup":
		runSetup(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Println("lmsgo " + version.Version)
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n\n", os.Args[1])
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}
