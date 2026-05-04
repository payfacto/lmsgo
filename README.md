# lmsgo

Delegate bulk I/O from Claude Code to a local [LM Studio](https://lmstudio.ai) model.
Single Go binary — no Python, no venv, no dependencies.

**Core principle: Claude = thinking. LM Studio = I/O.**

## Commands

```
lmsgo ask     --question "..." <file|dir> [...]   # bulk file reading
lmsgo write   --spec "..." --target <out> [ctx]   # boilerplate generation
lmsgo extract [-o output] <session.jsonl>          # strip Claude Code session to text
```

## Setup

### 1. Install LM Studio

Download and install [LM Studio](https://lmstudio.ai).

### 2. Download a model

Open LM Studio → **Discover** tab → search `gemma-4-e2b-it` → download from Hugging Face.
This must be done through LM Studio — do not download weights manually.

> Any instruction-tuned model works. `gemma-4-e2b-it` is a fast, lightweight default.
> Larger models (e.g. `qwen2.5-coder-7b-instruct`) produce better output at the cost of more RAM.

### 3. Keep the server running

**Option 1 — lmstudio CLI daemon (recommended)**

```powershell
# Windows
irm https://lmstudio.ai/install.ps1 | iex
```
```bash
# Linux / Mac
curl -fsSL https://lmstudio.ai/install.sh | bash
```
```bash
lms daemon up
```

**Option 2 — Desktop auto-start**

Settings (`Ctrl+,`) → enable **"Run the LLM server on login"**.

Docs: [lmstudio.ai/docs/developer/core/headless](https://lmstudio.ai/docs/developer/core/headless#auto-server-start)

### 4. Verify the server

```bash
curl http://localhost:1234/v1/models
```

Note the `id` of your model — you need it for `LMS_MODEL`.

### 5. Install lmsgo

```bash
# Build from source
cd C:/claudecode/scripts/lmsgo
go build -o lmsgo.exe .
cp lmsgo.exe ~/bin/
```

### 6. Configure environment

Add to `~/.bashrc`:

```bash
export LMS_MODEL=gemma-4-e2b-it   # model id from step 4
# LMS_BASE_URL defaults to http://localhost:1234/v1
# LMS_API_KEY  defaults to lm-studio (LM Studio ignores it)
```

### 7. Smoke test

```bash
lmsgo ask --question "What files are in this project?" README.md
```

### 8. Wire up Claude Code

Paste `CLAUDE_MD_SNIPPET.md` into `~/.claude/CLAUDE.md`. Claude will then self-route to `lmsgo` automatically.

## Usage

```bash
# Answer a question across files
lmsgo ask --question "Where are JWT tokens validated?" src/auth/

# Specific files
lmsgo ask --question "What env vars does this app require?" main.go config.go

# Directory with glob filter
lmsgo ask --glob "*.java" --question "Where is the DB pool configured?" src/

# Generate a file
lmsgo write --spec "Go table-driven tests for UserService" \
            --target internal/user/service_test.go \
            internal/user/service.go

# Preview without writing
lmsgo write --spec "Dockerfile for a Go app" --target Dockerfile --dry-run

# Documentation update after a session
lmsgo extract ~/.claude/projects/my-project/session.jsonl -o /tmp/chat.txt
lmsgo ask --question "What doc updates are needed? Give exact edits." \
          /tmp/chat.txt docs/architecture.md
```

## Developer Guide

### Structure

```
lmsgo/
├── main.go       # subcommand routing and usage
├── client.go     # LM Studio HTTP client (stdlib only)
├── ask.go        # ask subcommand
├── write.go      # write subcommand
├── extract.go    # extract subcommand
├── go.mod
└── CLAUDE_MD_SNIPPET.md
```

### Build

```bash
go build -o lmsgo.exe .
```

No external dependencies — stdlib only. The binary is self-contained.

### Adding a subcommand

1. Create `mycommand.go` with `func runMyCommand(args []string)`.
2. Add a `case "mycommand": runMyCommand(os.Args[2:])` in `main.go`.
3. Rebuild and copy the binary.

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LMS_BASE_URL` | `http://localhost:1234/v1` | LM Studio API base URL |
| `LMS_MODEL` | `local-model` | Model ID from `/v1/models` |
| `LMS_API_KEY` | `lm-studio` | Ignored by LM Studio |

### Token savings (expected)

| Task | Before | After |
|------|--------|-------|
| 5 files × 400 lines | ~8,000 tokens | ~400 tokens |
| Doc update after session | ~5,000 tokens | ~200 tokens |
| Generate 200-line boilerplate | ~3,000 tokens | ~200 tokens (review only) |
