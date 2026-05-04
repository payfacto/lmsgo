package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func runExtract(args []string) {
	fs_ := flag.NewFlagSet("extract", flag.ExitOnError)
	output := fs_.String("o", "", "Output file path (default: stdout)")
	last := fs_.Int("last", 0, "Only include the last N exchanges (0 = all)")

	fs_.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: lmsgo extract [flags] <session.jsonl>")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs_.PrintDefaults()
	}

	if err := fs_.Parse(args); err != nil {
		os.Exit(1)
	}
	if fs_.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "error: session file path is required")
		fs_.Usage()
		os.Exit(1)
	}

	sessionPath := fs_.Arg(0)
	f, err := os.Open(sessionPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	exchanges := parseSession(f)

	if *last > 0 && len(exchanges) > *last {
		exchanges = exchanges[len(exchanges)-*last:]
	}

	out := strings.Join(exchanges, "\n\n---\n\n")

	if *output != "" {
		if err := os.WriteFile(*output, []byte(out), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "error: write output: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "[lmsgo extract] wrote %d exchanges to %s\n", len(exchanges), *output)
	} else {
		fmt.Println(out)
	}
}

// claudeRecord covers the two JSONL shapes Claude Code produces.
type claudeRecord struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message"`
	Content json.RawMessage `json:"content"`
	Role    string          `json:"role"`
}

type claudeMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func parseSession(f *os.File) []string {
	var exchanges []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024) // 10 MB lines

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var rec claudeRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}

		role, text := extractRoleText(rec)
		if role == "" || role == "system" || text == "" {
			continue
		}
		exchanges = append(exchanges, fmt.Sprintf("[%s]\n%s", strings.ToUpper(role), text))
	}
	return exchanges
}

func extractRoleText(rec claudeRecord) (role, text string) {
	// Shape 1: {"type":"user"|"assistant", "message": {"role":..., "content":...}}
	if rec.Type == "user" || rec.Type == "assistant" {
		var msg claudeMessage
		if err := json.Unmarshal(rec.Message, &msg); err == nil && msg.Role != "" {
			return msg.Role, contentToText(msg.Content)
		}
		// fallback: type is the role
		return rec.Type, contentToText(rec.Content)
	}

	// Shape 2: {"message": {"role":..., "content":...}}
	if len(rec.Message) > 0 {
		var msg claudeMessage
		if err := json.Unmarshal(rec.Message, &msg); err == nil {
			return msg.Role, contentToText(msg.Content)
		}
	}

	return "", ""
}

func contentToText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	// Try plain string first
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	// Try array of content blocks
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			parts = append(parts, b.Text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
