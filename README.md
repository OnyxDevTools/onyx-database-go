# Onyx Database Go SDK

Go SDK and CLIs for Onyx Database, mirroring the TypeScript client with a contract-first design. The repository already contains the stabilized contract, SDK surface, and CLI entry points; tasks fill in the implementations while keeping deterministic outputs and parity with the TS experience.

## Contract-first layout
- `contract/` — stable, stdlib-only API guarded by compliance tests and `STABILITY.md`.
- `onyx/` — SDK implementation: configuration resolution, HTTP client, auth/signing, query execution, CRUD/cascade/schema, documents and secrets endpoints.
- `internal/` — shared helpers kept out of the public API.
- `cmd/onyx-schema-go/` — schema CLI: validate, diff, normalize, get, publish.
- `cmd/onyx-gen-go/` — codegen CLI: Go structs/table consts/helpers from `onyx.schema.json` or the API.
- `examples/` — runnable samples and README snippets that compile under the `docs` build tag.
- `codex-tasks/` — task manifest (`tasks.yaml`) plus per-task guides.

## Install

Requires Go 1.22+. Install the CLIs into your `$(go env GOBIN)` (or `$(go env GOPATH)/bin`) from the repo root:

```bash
go install ./cmd/onyx-schema-go
go install ./cmd/onyx-gen-go
```

## Build

Compile everything to ensure the SDK and CLIs are healthy:

```bash
go build ./...
go vet ./...
go test ./...
go tool cover -func=coverage.out 
```

## Usage (CLIs)

Install globally (recommended once published):

```bash
go install github.com/OnyxDevTools/onyx-database-go/cmd/onyx-schema-go@latest
go install github.com/OnyxDevTools/onyx-database-go/cmd/onyx-gen-go@latest
```

Local install from the repo (ensure `$(go env GOBIN)` or `$(go env GOPATH)/bin` is on your PATH):

```bash
go install ./cmd/onyx-schema-go
go install ./cmd/onyx-gen-go
```

- Schema CLI:
  - Validate: `onyx-schema-go validate`
  - Diff: `onyx-schema-go diff`
  - Get (API): `onyx-schema-go get` (replaces your schema with what is in onyx) or use `--print` to pretty-print without writing to your local schema)
  - Publish (API): `onyx-schema-go publish` if valid, will publish your remote database with the local one

- Codegen CLI:
  - File source: `onyx-gen-go --out /examples/onyx/models.go --package model`
  - API source: `onyx-gen-go --source api`

If you prefer not to install, prefix commands with `go run ./cmd/<cli> ...`.

## Examples

The Go examples mirror the TypeScript examples and are gated by the `docs` build tag so they stay out of normal builds. They rely on `ONYX_DATABASE_ID`, `ONYX_DATABASE_BASE_URL`, `ONYX_DATABASE_API_KEY`, and `ONYX_DATABASE_API_SECRET` (see `examples/shared/config.go` for defaults you can override).

- Compile-check everything: `go test -tags docs ./examples/...`
- Run a specific sample: wrap the desired function in a tiny `main` and execute with the `docs` tag. Example:

```bash
cat > /tmp/run_example.go <<'EOF'
package main

import (
	"context"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/examples/query"
)

func main() {
	if err := query.Basic(context.Background()); err != nil {
		log.Fatal(err)
	}
}
EOF

go run -tags docs /tmp/run_example.go
```

## Test and coverage

CI enforces coverage thresholds. Run the suite from the repo root:

Quick check: `go test ./...`

Coverage run:
```bash
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```

## Initialize the client

The Go client resolves configuration in the same order as the TS client: explicit values > environment variables > config files. Use `ClearConfigCache` between tests to reset cached values.

### Explicit configuration
```go
ctx := context.Background()

db, err := onyx.Init(ctx, onyx.Config{
    DatabaseID:      "db_123",
    DatabaseBaseURL: "https://api.onyx.dev",
    APIKey:          os.Getenv("ONYX_DATABASE_API_KEY"),
    APISecret:       os.Getenv("ONYX_DATABASE_API_SECRET"),
    LogRequests:     true,  // optional: enable request logging
    LogResponses:    false, // optional: enable response logging
})
if err != nil { log.Fatal(err) }
```

### Resolver, env vars, and config files
```go
// Environment variables are read when explicit values are empty
os.Setenv("ONYX_DATABASE_ID", "db_123")
os.Setenv("ONYX_DATABASE_BASE_URL", "https://api.onyx.dev")
os.Setenv("ONYX_DATABASE_API_KEY", "key_abc")
os.Setenv("ONYX_DATABASE_API_SECRET", "secret_xyz")

ctx := context.Background()
db, err := onyx.InitWithDatabaseID(ctx, "db_123")
if err != nil { log.Fatal(err) }

defer onyx.ClearConfigCache() // reset between test cases
```

You can also supply a JSON file (lowest precedence) named `onyx-database.json`, `onyx-database-<dbid>.json`, or pointed to via `ONYX_CONFIG_PATH`:
```json
{
  "databaseId": "db_123",
  "databaseBaseUrl": "https://api.onyx.dev",
  "apiKey": "key_abc",
  "apiSecret": "secret_xyz"
}
```

## Query builder parity

The fluent query API mirrors the TS builder, supporting filters, projections, nested resolves, sorting, pagination, and streaming.

```go
ctx := context.Background()
client := mustInit(ctx) // helper that wraps onyx.Init

page, err := db.From("User").
    Select("id", "email", "profile.avatarUrl").
    Where(onyx.Contains("email", "@example.com")).
    And(onyx.Gte("createdAt", "2024-01-01T00:00:00Z")).
    Resolve("roles", "profile").
    OrderBy(onyx.Desc("createdAt")).
    Limit(25).
    Page(ctx, "") // empty cursor = first page
if err != nil { log.Fatal(err) }

for _, user := range page.Items {
    fmt.Println(user)
}

iter, err := db.From("Event").Stream(ctx)
if err != nil { log.Fatal(err) }
for iter.Next() {
    evt := iter.Value()
    // handle event incrementally
}
if err := iter.Err(); err != nil { log.Fatal(err) }
```

## Resolves and nested data

Use `Resolve` with dotted paths to hydrate related records just like the TS client:
```go
post, err := db.From("Post").
    Where(onyx.Eq("id", "post_123")).
    Resolve("author.profile", "comments.author").
    Limit(1).
    List(ctx)
if err != nil { log.Fatal(err) }
```

## Cascades

Cascade specs match TS semantics and are constructed via `onyx.Cascade`:
```go
spec := onyx.Cascade("userRoles:UserRole(userId,id)")

if err := db.Cascade(spec).Save(ctx, "User", user); err != nil {
    log.Fatal(err)
}

if err := db.Cascade(spec).Delete(ctx, "User", user.ID); err != nil {
    log.Fatal(err)
}
```

## Documents and secrets APIs

The Go SDK mirrors the TS helpers for auxiliary APIs as well:

```go
ctx := context.Background()
client := mustInit(ctx)

doc, err := db.Documents().Get(ctx, "doc_123")
if err != nil { log.Fatal(err) }

doc.Data["status"] = "updated"
if _, err := db.Documents().Save(ctx, doc); err != nil {
    log.Fatal(err)
}

secret := onyx.Secret{Key: "API_KEY", Value: "abc"}
if _, err := db.PutSecret(ctx, secret); err != nil {
    log.Fatal(err)
}
```

## Schema CLI (onyx-schema-go)

Validate and normalize contract-first schemas locally, or fetch/publish them via the API:
```bash
onyx-schema-go validate --schema ./onyx.schema.json
onyx-schema-go diff --a ./onyx.schema.json --b ./next.schema.json --json
onyx-schema-go normalize --schema ./onyx.schema.json --out ./onyx.normalized.json
onyx-schema-go get --database-id db_123 --out ./onyx.schema.json
# or just print the remote schema using your env/config (onyx-database.json):
onyx-schema-go get --print
onyx-schema-go publish --schema ./onyx.schema.json
```

If `--database-id` is omitted, the CLI resolves credentials from env vars or config files such as `./onyx-database.json` (also `~/.onyx/onyx-database.json`).

## Codegen CLI (onyx-gen-go)

Generate strongly-typed models, table constants, and helpers with deterministic output that only import `contract` (plus `time` when timestamps are requested):
```bash
onyx-gen-go --schema ./onyx.schema.json --out ./models/onyx.go --package models --tables User,Order --timestamps time

# Pull the schema from the API and emit the same helpers
onyx-gen-go --source api --database-id db_123 --out ./models/onyx.go --package models
```

Generated helpers let you write `q := FromUser(client).Where(onyx.Eq("id", userID))` while keeping the public API aligned with the TS db.

## Working with the task pack
- Tasks are dependency-aware: P0 setup → P1 contract freeze → P2 CLIs → P3 SDK core → P4 docs/CI/release/parity.
- Start with `P1_CONTRACT_10_FREEZE_V1` before SDK work; P2/P3 tasks can run in parallel when dependencies allow.
- Use the per-task markdown files in `codex-tasks/tasks/` for acceptance criteria and implementation notes.
