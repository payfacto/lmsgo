# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build

```bash
# Dev build (version = "dev")
go build -o lmsgo.exe .

# Release build with version from tag
VERSION=$(git describe --tags --always --dirty)
go build -ldflags "-X 'github.com/payfacto/lmsgo/internal/version.Version=${VERSION}'" -o lmsgo.exe .
```

No external dependencies — stdlib only (`go 1.22`). After building, copy the binary to a directory on `PATH`:

```bash
cp lmsgo.exe ~/bin/
```

## Architecture

Single `main` package with one file per subcommand:

- [main.go](main.go) — subcommand dispatch (`os.Args[1]` switch); imports `internal/version`
- [client.go](client.go) — `complete()` function: the only path to the LM Studio API (`POST /v1/chat/completions`). Reads `LMS_BASE_URL`, `LMS_MODEL`, `LMS_API_KEY` env vars.
- [ask.go](ask.go) — `runAsk`: walks files/dirs into an XML corpus, sends a three-message conversation (system + corpus + question)
- [write.go](write.go) — `runWrite`: builds a generation prompt from `--spec` and context files, writes result directly to `--target`
- [extract.go](extract.go) — `runExtract`: parses Claude Code JSONL session files (two different record shapes) into readable `[ROLE]\ntext` exchanges
- [internal/version/version.go](internal/version/version.go) — `Version` var, defaulting to `"dev"`, injected at build time via `-ldflags`

## Adding a Subcommand

1. Create `mycommand.go` with `func runMyCommand(args []string)`.
2. Add `case "mycommand": runMyCommand(os.Args[2:])` to the switch in [main.go](main.go).
3. Rebuild.

## Environment Variables

| Variable | Default | Purpose |
|---|---|---|
| `LMS_BASE_URL` | `http://localhost:1234/v1` | LM Studio API base |
| `LMS_MODEL` | `local-model` | Model ID from `/v1/models` |
| `LMS_API_KEY` | `lm-studio` | Ignored by LM Studio |

## LM Studio Prerequisite

The binary requires a running LM Studio server. Start it with `lms daemon up` or enable auto-start in LM Studio settings. Verify with:

```bash
curl http://localhost:1234/v1/models
```

## Versioning and Releases

**Source of truth is Bitbucket. GitHub is CI/release only. Never push directly to the GitHub mirror.**

### Release flow

Tag on `main` in Bitbucket and push:

```bash
git tag v1.2.3
git push origin v1.2.3
```

- [bitbucket-pipelines.yml](bitbucket-pipelines.yml) mirrors `main` + tags to `github.com/payfacto/lmsgo`.
- [.github/workflows/release.yml](.github/workflows/release.yml) fires on `v*` tags, builds linux/windows/darwin, creates a GitHub Release, and updates the `payfacto/homebrew-tap` formula.

### One-time setup

- **Bitbucket**: `Repository settings → Pipelines → SSH keys → Generate keys`. Add the public key to GitHub as a deploy key (write access). Add `github.com` to known hosts.
- **GitHub secret**: Create a fine-grained PAT with `Contents: write` on `payfacto/homebrew-tap` and add it as `HOMEBREW_TAP_TOKEN` in the GitHub repo secrets.

### Tagging rules

- Always use semver (`vMAJOR.MINOR.PATCH`), tag on `main` only.
- Never reuse a tag that has already been released.
