# Contributing

Thanks for helping improve the Onyx Database Go SDK. This guide covers the local workflow: regenerating examples, running tests, and using the CLIs from both the root module and the examples module.

## Prerequisites
- Go 1.22.x (we standardize on 1.22.8). The repo assumes Go is provided via gvm.
  - Install gvm: `bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)`
  - Install toolchain: `gvm install go1.22.8` (and any other versions you need).
  - Set the project default: `gvm use go1.22.8 --default` (or run `gvm use go1.22.8` in this repo before builds/tests).
- No sandboxed network installs are required; CLIs build from source once gvm is set.

## Project layout basics
- Root module: SDK, CLIs, contract, implementation (`go.mod` at repo root).
- Examples module: lives in `examples/`, has its own `go.mod`, generated client in `examples/gen/onyx`, and sample config/schema (`examples/config/onyx-database.json`, `examples/api/onyx.schema.json`).
- Code generation entrypoint: `generate.go` at repo root (runs the generator against `examples/api/onyx.schema.json` and writes `examples/gen/onyx` deterministically).

## Regenerating the example client
From the examples:
```bash
go generate   # runs onyx-gen-go against examples/api/onyx.schema.json and writes examples/gen/onyx
```
The generation uses a fixed `ONYX_GEN_TIMESTAMP` for deterministic headers. If you change the schema under `examples/api/onyx.schema.json`, rerun `go generate ./...` and commit the updated files under `examples/gen/onyx`.

## Building and testing
From the repo root:
```bash
go build ./...
go vet ./...
go test ./...
```
## See Code Coverage:
go test ./... -coverprofile=coverage.out -covermode=atomic

## Using the CLIs locally
Build/install once from the repo root:
```bash
cd examples
go install ../cmd/onyx-go
onyx-go gen init
go generate $ or onyx-go gen
../scripts/run-examples.sh
```


after you've geneted the go stubs there is now a 
```bash
go generate
```

## Editor setup (VS Code)
- The repo includes `.vscode/settings.json` that pins the Go toolchain for a smooth dev flow:
  - `GOROOT`/`PATH` set to `/Users/cosborn/.gvm/gos/go1.22.8`
  - `GOTOOLCHAIN=go1.22.8` so lint/test/build use Go 1.22.8 without manual exports
- Launch/tasks already use these settings; open the folder in VS Code and run/debug without extra env exports.

Schema CLI quick checks (root or examples, adjust paths as needed):
```bash
onyx-go schema info
onyx-go schema get --print
onyx-go schema validate
onyx-go schema diff
```
Defaults look in `./config/onyx-database.json` for credentials and `./api/onyx.schema.json` for schema files (matching generator defaults in the examples module).

### CLI reference

`onyx-go gen`
- Flags: `--schema ./api/onyx.schema.json`, `--source file`, `--database-id ""` (only when `--source=api`), `--out ./gen/onyx`, `--package ""` (defaults to `onyx`), `--tables ""` (all tables), `--timestamps time`
- Notes: Generates the typed client into `--out`; defaults assume an `api/` + `gen/` layout.

`onyx-go gen init`
- Flags: `--file generate.go`, `--schema ./api/onyx.schema.json`, `--source file`, `--database-id ""`, `--out ./gen/onyx`, `--package onyx`, `--tables ""`, `--timestamps time`
- Notes: Writes a go:generate anchor; subsequent `go generate` uses these defaults.

`onyx-go schema info`
- Flags: `--database-id ""`, `--config ""`, `--no-verify` (false)
- Notes: Resolves config (env → `ONYX_CONFIG_PATH` → `config/onyx-database.json` → home). Verifies connectivity unless `--no-verify`.

`onyx-go schema get`
- Flags: `--database-id ""`, `--out api/onyx.schema.json`, `--print` (false)
- Notes: Fetches schema from API, normalizes, and writes to `--out` (creates parent dirs). With `--print`, writes to stdout and skips file output.

`onyx-go schema validate`
- Flags: `--schema api/onyx.schema.json`
- Notes: Parses + validates a local schema file.

`onyx-go schema normalize`
- Flags: `--schema api/onyx.schema.json`, `--out ""`
- Notes: Normalizes a schema; writes to stdout when `--out` is empty, otherwise writes file.

`onyx-go schema diff`
- Flags: `--a api/onyx.schema.json`, `--b ""`, `--database-id ""`, `--json` (false)
- Notes: Compares two schemas. If `--b` is empty, fetches updated schema from API (optionally for `--database-id`; otherwise uses resolved credentials).

`onyx-go schema publish`
- Flags: `--database-id ""`, `--schema api/onyx.schema.json`
- Notes: Normalizes and publishes the local schema to the API.

Generator CLI (examples module paths):
```bash
onyx-go gen --schema ./examples/api/onyx.schema.json --out ./examples/gen/onyx --package onyx
# or from the examples module:
onyx-go gen --schema ./api/onyx.schema.json --out ./gen/onyx --package onyx
```

## Linting
- We pin linting to Go 1.22.8. The repo’s `.golangci.yml` sets `GOROOT`, `PATH`, and `GOTOOLCHAIN` for you; run from the repo root:
  ```bash
  golangci-lint run
  ```
- If you don’t want to change your global shell or gvm state, install the pinned linter into this repo’s `./bin` with:
  ```bash
  scripts/install-local-tools.sh
  ```
  The VS Code settings and integrated terminal are configured to prefer `./bin` and Go 1.22.8 automatically.
- If you still hit toolchain/version errors, reinstall golangci-lint with Go 1.22.8 (e.g., `GOROOT=/Users/cosborn/.gvm/gos/go1.22.8 PATH="$GOROOT/bin:$PATH" go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`) so the binary matches the pinned toolchain.

## Releasing (bump-version.sh)

The repo ships an interactive release helper at `scripts/bump-version.sh` that:
- Verifies a clean tree on `main`
- Runs `go mod tidy` validation, tests, lint, build, and a smoke example tests
- Prompts for semver bump (patch/minor/major) and a release message
- Computes the next tag (first release starts at `v0.0.1`)
- Commits (chore(release): …), tags, and pushes to origin; CI publishes on tag push

Make sure your working tree is clean (commit/stash everything) before running the script; it will abort on uncommitted changes.
Usage from the repo root:
```bash
scripts/bump-version.sh
```
