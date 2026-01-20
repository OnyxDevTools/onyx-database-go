# AGENTS

## Quick orientation
- Go SDK + CLIs for Onyx Database. Root module hosts packages `contract` (public API types), `impl` (runtime), `onyx` (thin re-export/wiring). `examples/` is a separate module with generated client + runnable samples.
- CLIs live under `cmd/`: `onyx-go` (gen + schema entrypoint), `onyx-gen-go` (generator), `onyx-schema-go` (schema tools). Helper scripts in `scripts/`.
- `contract` must stay stdlib-only and stable; see `contract/STABILITY.md` before altering exported surfaces. `onyx` should remain a light facade over `impl`.

## Tooling & build
- Go 1.22.x pinned to 1.22.8 via gvm; env baked into `.vscode/settings.json` and `.golangci.yml`.
- Run from repo root: `go build ./...`, `go vet ./...`, `go test ./... -coverprofile=coverage.out -covermode=atomic`; `golangci-lint run` if available (install via `scripts/install-local-tools.sh`).
- Format with `gofmt`; avoid adding dependencies, especially under `contract`.
- `examples/` has its own `go.mod`; run its tests/commands from inside that folder if you touch it.

## Config & auth
- Resolver precedence: explicit `onyx.Config` > env (`ONYX_DATABASE_ID`, `ONYX_DATABASE_BASE_URL`, `ONYX_DATABASE_API_KEY`, `ONYX_DATABASE_API_SECRET`) > files. `ConfigPath` or `ONYX_CONFIG_PATH` overrides search.
- Default file search: `config/onyx-database-<id>.json`, `config/onyx-database.json`, `./onyx-database-<id>.json`, `./onyx-database.json`, `~/.onyx/onyx-database-<id>.json`, `~/.onyx/onyx-database.json`, `~/onyx-database*.json` (partition is read, too).
- Cache TTL defaults to 5m; `onyx.ClearConfigCache()` clears resolver + HTTP client caches. `Config.Partition` seeds queries and deletes; `Config.HTTPClient`, `Clock`, and `Sleep` are injectable for tests.
- `ONYX_DEBUG=true` forces request/response logging regardless of `Config.LogRequests`/`LogResponses`.
- AI base URL defaults to `https://ai.onyx.dev`; override with `Config.AIBaseURL` or `ONYX_AI_BASE_URL` while reusing the same API key/secret.

## Runtime behavior & invariants
- Query builder is immutable (methods clone state). `InPartition("")` clears partition; otherwise config partition propagates to queries and delete calls.
- HTTP client signs with `x-onyx-key/secret`, optional logging, and caches per baseURL + signer; clear cache when credentials change to avoid stale clients.
- `batchSave` chunks at 500 by default and retries once after 50ms on 429/502/503/504; preserve this behavior when modifying batch logic.
- Schema APIs handle legacy shapes; `contract.NormalizeSchema` sorts tables/fields/resolvers/indexes/triggers for deterministic operations. Publish via `/schemas/{db}?publish=true|false`; validate via `/schemas/{db}/validate`.
- Errors are `*contract.Error` with deterministic formatting and `Meta["status"]`; streaming iterator scans newline-delimited JSON up to 10MB buffer.

## Code generation & examples
- `go generate` (root) runs `cmd/onyx-go gen` against `examples/api/onyx.schema.json` and writes `examples/gen/onyx` with `ONYX_GEN_TIMESTAMP` pinned for deterministic headers. Regenerate instead of manual edits when schema or generator changes.
- Generator usage: `onyx-go gen --schema ./api/onyx.schema.json --out ./gen/onyx --package onyx` (add `--source api --database-id ...` to pull from API). `onyx-go gen init` writes a go:generate anchor.
- Schema CLI: `onyx-go schema {info|get|validate|diff|publish}` uses the same resolver; `--print` avoids file output.
- Examples rely on `examples/config/onyx-database.json`; `scripts/run-examples.sh` executes each sample via `go run` and expects the `example: completed` marker.

## Example writing guidelines
- Audience-first: every example should read like a minimal tutorial, explaining intent with clear variable names and `log.Fatal`/`log.Printf`-style errors instead of test assertions. Prefer printing meaningful output (e.g., marshaled results) so users see what Onyx returns.
- Logic-driven failures: examples should fail only when logic or SDK calls fail (e.g., nil checks, unexpected counts), not via testing frameworks. Keep setup minimal and reuse the resolver/env config defaults.
- Pass/Fail contract: `scripts/run-examples.sh` marks PASS only when `go run` exits 0 and stdout contains `example: completed`; include that final log line on success.
- Ergonomics: keep SDK usage to a single import where possible (std lib + one generated/SDK import). Avoid refactors that introduce extra SDK packages into examples.
- Backward compatibility: do not change existing example code to use new contracts or signatures—additive examples only. Any API surface changes must remain backward compatible for users upgrading the SDK.

## Testing notes
- Tests lean on stdlib; use provided hooks (e.g., resolver/env fakes, HTTP client/test hooks) for determinism instead of reaching into globals.
- Update both contract and impl tests when changing public structures or schema normalization. Keep generated files updated if they are tracked.
