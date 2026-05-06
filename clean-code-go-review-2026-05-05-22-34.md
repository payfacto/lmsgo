# clean-code:go Review — 2026-05-05 22:34

**Files reviewed:** `write.go`, `client.go` (changes made this session)
**Total violations:** 3 (all fixed)

## Summary
| Category | Count |
|----------|-------|
| Comments | 0 |
| Functions | 0 |
| General | 3 |
| Go-Specific | 0 |
| Names | 0 |
| Tests | 0 (existing — flagged as gap) |

## Tasks
- [x] `write.go:104` — **G9** Removed speculative `target` parameter from `stripPreamble` and the `_ = target` discard. Restore only when a per-extension heuristic actually exists.
- [x] `write.go:105-109` — **G5/perf** Hoisted the `keywords` slice to a package-level `codeStartMarkers` var so it isn't rebuilt on every call. Renamed for clarity.
- [x] `write.go:117-123` — **G16** Extracted inline byte-range word check to a named `isWordByte` helper with a doc comment explaining the role.

## Verification
- `go build ./...` → success
- `go vet ./...` → no issues
- `lmsgo write` smoke test → byte-identical clean output to pre-refactor

## Open gaps (not fixed in this pass)
- **No tests in repo.** `stripPreamble` is a pure function ideal for table-driven tests covering: clean input, leading planning preamble, content-then-fence-then-duplicate, no markers found, trailing whitespace normalization. Worth a follow-up issue. (T1, T5)
- **`runWrite` and `runAsk` size.** Each handler does parsing, validation, prompt building, LLM call, and output. Acceptable for CLI dispatchers; flagged for awareness if the file grows further. (G30)
- **Duplicated HTTP plumbing in `client.go`.** `complete` and `listModels` both build a request, set headers, dispatch, read body, unmarshal. Different verbs and bodies, so extracting a helper would force awkward generic types. Left flat — preferable to the wrong abstraction. (G5 considered, declined)
