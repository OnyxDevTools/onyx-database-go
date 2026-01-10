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

Schema CLI quick checks (root or examples, adjust paths as needed):
```bash
onyx-go schema info
onyx-go schema get --print
onyx-go schema validate
onyx-go schema diff
```

Generator CLI (examples module paths):
```bash
onyx-go gen --schema ./examples/api/onyx.schema.json --out ./examples/gen/onyx --package onyx
# or from the examples module:
onyx-go gen --schema ./onyx.schema.json --out ./gen/onyx --package onyx
```

## Releasing (bump-version.sh)

The repo ships an interactive release helper at `scripts/bump-version.sh` that:
- Verifies a clean tree on `main`
- Runs `go mod tidy` validation, tests, lint, build, and a smoke example
- Prompts for semver bump (patch/minor/major) and a release message
- Computes the next tag (first release starts at `v0.0.1`)
- Commits (chore(release): …), tags, and pushes to origin; CI publishes on tag push

Make sure your working tree is clean (commit/stash everything) before running the script; it will abort on uncommitted changes.
Usage from the repo root:
```bash
scripts/bump-version.sh
```

