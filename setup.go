package main

import (
	_ "embed"
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

//go:embed CLAUDE_MD_SNIPPET.md
var snippetContent string

const (
	sentinelStart = "<!-- lmsgo:start -->"
	sentinelEnd   = "<!-- lmsgo:end -->"
)

func runSetup(args []string) {
	fs_ := flag.NewFlagSet("setup", flag.ExitOnError)
	modelFlag := fs_.String("model", "", "Model ID to use (skips interactive selection)")
	dryRun := fs_.Bool("dry-run", false, "Preview changes without writing anything")

	fs_.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: lmsgo setup [flags]")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs_.PrintDefaults()
	}

	if err := fs_.Parse(args); err != nil {
		os.Exit(1)
	}

	// Step 1: detect LM Studio and available models
	fmt.Fprintln(os.Stderr, "[1/3] Checking LM Studio...")
	models, err := listModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "      ! Could not reach LM Studio: %v\n", err)
		fmt.Fprintln(os.Stderr, "        Start it with: lms daemon up")
		fmt.Fprintln(os.Stderr, "        Continuing with placeholder model name.")
	} else {
		fmt.Fprintf(os.Stderr, "      + Reachable — %d model(s) loaded\n", len(models))
	}

	// Step 2: resolve model
	selectedModel := *modelFlag
	if selectedModel == "" {
		switch len(models) {
		case 0:
			selectedModel = defaultModel
			fmt.Fprintf(os.Stderr, "      Using default: %s\n", selectedModel)
		case 1:
			selectedModel = models[0]
			fmt.Fprintf(os.Stderr, "      Auto-selected: %s\n", selectedModel)
		default:
			selectedModel = promptModelSelection(models)
		}
	}

	// Step 3: env file
	fmt.Fprintln(os.Stderr, "[2/3] Environment variables...")
	writeEnvFile(selectedModel, *dryRun)

	// Step 4: patch CLAUDE.md
	fmt.Fprintln(os.Stderr, "[3/3] Patching ~/.claude/CLAUDE.md...")
	patchClaudeMD(*dryRun)

	if !*dryRun {
		fmt.Fprintln(os.Stderr, "\nSetup complete.")
	}
}

func promptModelSelection(models []string) string {
	fmt.Fprintln(os.Stderr, "      Available models:")
	for i, m := range models {
		fmt.Fprintf(os.Stderr, "        [%d] %s\n", i+1, m)
	}
	fmt.Fprint(os.Stderr, "      Select [1]: ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return models[0]
		}
		if n, err := strconv.Atoi(line); err == nil && n >= 1 && n <= len(models) {
			return models[n-1]
		}
		return line // treat as a literal model ID
	}
	return models[0]
}

// envFilePath returns the path to write env vars and the shell expression to source it.
func envFilePath() (path, sourceExpr string) {
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			home, _ := os.UserHomeDir()
			appdata = filepath.Join(home, "AppData", "Roaming")
		}
		p := filepath.Join(appdata, "lmsgo", "env.ps1")
		return p, fmt.Sprintf(`. "%s"`, p)
	}
	home, _ := os.UserHomeDir()
	p := filepath.Join(home, ".config", "lmsgo", "env")
	return p, fmt.Sprintf(". %s", p)
}

func envFileContents(model string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(
			"$env:LMS_BASE_URL = \"%s\"\n$env:LMS_MODEL     = \"%s\"\n$env:LMS_API_KEY  = \"%s\"\n",
			defaultBaseURL, model, defaultAPIKey,
		)
	}
	return fmt.Sprintf(
		"export LMS_BASE_URL=%s\nexport LMS_MODEL=%s\nexport LMS_API_KEY=%s\n",
		defaultBaseURL, model, defaultAPIKey,
	)
}

func writeEnvFile(model string, dryRun bool) {
	path, sourceExpr := envFilePath()
	contents := envFileContents(model)

	if dryRun {
		fmt.Fprintf(os.Stderr, "      [dry-run] would write %s:\n", path)
		for _, line := range strings.Split(strings.TrimSpace(contents), "\n") {
			fmt.Fprintf(os.Stderr, "        %s\n", line)
		}
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "      ! Could not create directory: %v\n", err)
		return
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "      ! Could not write env file: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "      + Written to %s\n", path)

	if runtime.GOOS == "windows" {
		fmt.Fprintf(os.Stderr, "        Add to your PowerShell profile ($PROFILE):\n")
	} else {
		fmt.Fprintf(os.Stderr, "        Add to ~/.bashrc or ~/.zshrc:\n")
	}
	fmt.Fprintf(os.Stderr, "          %s\n", sourceExpr)
}

func patchClaudeMD(dryRun bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "      ! Could not find home directory: %v\n", err)
		return
	}
	claudePath := filepath.Join(home, ".claude", "CLAUDE.md")

	existing, err := os.ReadFile(claudePath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "      ! Could not read %s: %v\n", claudePath, err)
		return
	}

	if strings.Contains(string(existing), sentinelStart) {
		fmt.Fprintln(os.Stderr, "      + Already present — skipping")
		return
	}

	addition := "\n" + sentinelStart + "\n" + strings.TrimSpace(snippetContent) + "\n" + sentinelEnd + "\n"

	if dryRun {
		fmt.Fprintf(os.Stderr, "      [dry-run] would append to %s:\n", claudePath)
		for _, line := range strings.Split(strings.TrimSpace(addition), "\n") {
			fmt.Fprintf(os.Stderr, "        %s\n", line)
		}
		return
	}

	f, err := os.OpenFile(claudePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "      ! Could not open %s: %v\n", claudePath, err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(addition); err != nil {
		fmt.Fprintf(os.Stderr, "      ! Could not write to %s: %v\n", claudePath, err)
		return
	}

	fmt.Fprintf(os.Stderr, "      + Snippet appended to %s\n", claudePath)
}
