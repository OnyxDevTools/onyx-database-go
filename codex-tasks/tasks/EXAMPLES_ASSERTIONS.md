Title: Harden examples with assertions and consistent error handling

Goal: Keep examples user-friendly but make them self-checking so running `go test ./...` exercises the example flows and catches regressions.

Scope
- All example command mains under `examples/**/cmd/**/main.go` (save/query/delete/stream/schema/secrets/etc.).
- Do not convert to full unit tests; keep them runnable snippets for users.

Plan
- Add a tiny shared helper (e.g., `examples/assert/assert.go`) with concise utilities:
  - `Must(err error)` to fail fast.
  - `Require(cond bool, msg string, args ...any)` for sanity checks.
- Update examples to:
  - Replace `log.Fatal`/`panic` with `assert.Must`.
  - Add minimal sanity checks per example (e.g., non-empty saved ID, expected field present, page cursor handled) that do not rely on specific remote data beyond what the example seeds/creates.
  - Keep outputs readable for learners; no verbose test harness.
- Ensure imports stay minimal; avoid pulling in testing packages in mains.
- Run `go test ./...` to verify no regressions after updates.

Notes
- Keep contract/API boundaries unchanged.
- Avoid hardcoding environment-sensitive expectations (e.g., exact record counts from remote API); focus on what the example itself produces or on structurally valid responses.
