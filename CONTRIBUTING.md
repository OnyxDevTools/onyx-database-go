# Contributing

Thanks for helping improve the Onyx Database Go SDK. This guide covers the local workflow: regenerating examples, running tests, and using the CLIs from both the root module and the examples module.

## Prerequisites
- Go 1.22 or newer installed and on your PATH.
- No sandboxed network installs are required; CLIs build from source.

## Project layout basics
- Root module: SDK, CLIs, contract, implementation (`go.mod` at repo root).
- Examples module: lives in `examples/`, has its own `go.mod`, generated client in `examples/gen/onyx`, and sample config/schema (`examples/config/onyx-database.json`, `examples/api/onyx.schema.json`).
- Code generation entrypoint: `generate.go` at repo root (runs the generator against `examples/api/onyx.schema.json` and writes `examples/gen/onyx` deterministically).

## Regenerating the example client
From the examples:
```bash
go generate ./...   # runs onyx-gen-go against examples/api/onyx.schema.json and writes examples/gen/onyx
```
The generation uses a fixed `ONYX_GEN_TIMESTAMP` for deterministic headers. If you change the schema under `examples/api/onyx.schema.json`, rerun `go generate ./...` and commit the updated files under `examples/gen/onyx`.

## Building and testing
From the repo root:
```bash
go build ./...
go vet ./...
go test ./...
```

## Using the CLIs locally
Build/install once from the repo root:
```bash
cd examples
go install ../cmd/onyx-go
onyx-go gen init
onyx-go gen
../scripts/run-examples.sh
```

after you've geneted the go stubs there is now a 
```bash
go generate
```



Schema CLI quick checks (root or examples, adjust paths as needed):
```bash
onyx-schema-go validate --schema ./examples/api/onyx.schema.json
onyx-schema-go diff --a ./examples/api/onyx.schema.json --b ./next.schema.json --json
onyx-schema-go get --database-id "$ONYX_DATABASE_ID" --out ./examples/api/onyx.schema.json
```

Generator CLI (examples module paths):
```bash
onyx-gen-go --schema ./examples/api/onyx.schema.json --out ./examples/gen/onyx --package onyx
# or from the examples module:
cd examples
onyx-gen-go --schema ./onyx.schema.json --out ./gen/onyx --package onyx
```

## Working inside the examples module
- Use the sample config at `examples/config/onyx-database.json` or set `ONYX_DATABASE_ID`, `ONYX_DATABASE_BASE_URL`, `ONYX_DATABASE_API_KEY`, `ONYX_DATABASE_API_SECRET`.
- Run individual samples with `cd examples && go run ./query/cmd/list` (or any other path under `examples/...`).

## go:generate scaffold for your own project
In your own app (not this repo), you can add a helper like:
```go
// internal/codegen/gen.go
package codegen

//go:generate onyx-gen-go --schema ../../onyx.schema.json --out ../../gen/onyx --package onyx
//go:generate gofmt -w ../../gen/onyx
```
Then run `go generate ./internal/codegen` from your repo root to refresh your generated client.

## Before sending changes
1. Run `go generate ./...` from the repo root (commit any changes under `examples/gen/onyx`).
2. Run `go vet ./...` and `go test ./...`.
3. `cd examples && go test ./...`.
4. Optionally `./scripts/run-examples.sh` to exercise the binaries end-to-end.

