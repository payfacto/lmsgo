## LM Studio Delegation (Token Saving)

**Rule: Claude = thinking. LM Studio = I/O.**

Use `lmsgo ask` instead of reading files yourself when: 3+ files, or any single file >400 lines.
Use `lmsgo write` for: test files, config scaffolding, doc drafts, repetitive boilerplate.
Use `lmsgo extract` + `lmsgo ask` for documentation updates after a feature session — never write docs directly.

```bash
lmsgo ask --question "..." file1 file2
lmsgo ask --glob "*.go" --question "..." src/
lmsgo write --spec "..." --target output.go context.go
lmsgo extract -o /tmp/chat.txt ~/.claude/projects/<slug>/session.jsonl
```

**Flag ordering**: all flags must come BEFORE positional paths. Go's `flag` package stops parsing at the first non-flag, so `lmsgo extract file.jsonl --last 5` silently ignores `--last`.

**On `API error: Context length exceeded`**: do NOT retry — the prompt is too big for the model's context. Tighten `--glob`, split the corpus into smaller batches, or fall back to direct `Read`/`Grep`.

Do NOT delegate: tasks <2000 tokens, debugging, architecture decisions, security code, anything needing exact line numbers.
