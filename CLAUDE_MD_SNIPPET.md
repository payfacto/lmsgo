## LM Studio Delegation (Token Saving)

**Rule: Claude = thinking. LM Studio = I/O.**

Use `lmsgo ask` instead of reading files yourself when: 3+ files, or any single file >400 lines.
Use `lmsgo write` for: test files, config scaffolding, doc drafts, repetitive boilerplate.
Use `lmsgo extract` + `lmsgo ask` for documentation updates after a feature session — never write docs directly.

```bash
lmsgo ask --question "..." file1 file2
lmsgo write --spec "..." --target output.go context.go
lmsgo extract ~/.claude/projects/<slug>/session.jsonl -o /tmp/chat.txt
```

Do NOT delegate: tasks <2000 tokens, debugging, architecture decisions, security code, anything needing exact line numbers.
