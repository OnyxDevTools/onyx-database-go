# Onyx Database Go SDK

Go SDK and CLIs for Onyx Database, mirroring the TypeScript client with a contract-first design. The repository already contains the stabilized contract, SDK surface, and CLI entry points; tasks fill in the implementations while keeping deterministic outputs and parity with the TS experience.

## Contract-first layout
- `contract/` — stable, stdlib-only API guarded by compliance tests and `STABILITY.md`.
- `onyx/` — SDK implementation: configuration resolution, HTTP client, auth/signing, query execution, CRUD/cascade/schema, documents and secrets endpoints.
- `internal/` — shared helpers kept out of the public API.
- `cmd/onyx-schema-go/` — schema CLI: validate, diff, normalize, get, publish.
- `cmd/onyx-gen-go/` — codegen CLI: Go structs/table consts/helpers from `onyx.schema.json` or the API.
- `examples/` — runnable samples and README snippets that build as normal binaries.
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
- File source: `onyx-gen-go --out ./gen/onyx --package onyx`

- Repository structure and purpose:

```
├── AGENTS.md                      # Codex-specific guidance for this repo
├── cmd                            # Command-line tools
│   ├── onyx-gen-go                # Go client/code generator
│   │   ├── doc.go                 # Package docs
│   │   ├── generator              # Generator logic and tests
│   │   ├── main.go                # CLI entrypoint
│   │   └── main_test.go           # CLI wiring tests
│   └── onyx-schema-go             # Schema CLI (diff/validate/get/publish)
│       ├── commands               # Subcommand implementations and tests
│       ├── doc.go                 # Package docs
│       └── main.go                # CLI entrypoint
├── contract                       # Public contract types and helpers (stdlib-only)
├── examples                       # End-to-end usage examples (query/save/schema/secrets/streams)
│   ├── cmd                        # Misc example entrypoints (e.g., seeding)
│   ├── delete/query/...           # Delete/query examples
│   ├── document                   # Document CRUD examples
│   ├── query                      # Query examples (filters, paging, aggregates, etc.)
│   ├── save                       # Save/batch/cascade examples
│   ├── schema                     # Schema client examples
│   ├── secrets                    # Secrets client examples
│   └── stream                     # Streaming examples (create/update/delete/listen)
├── gen/onyx                       # Generated typed client/models (deterministic output)
├── impl                           # Internal implementation of the Onyx client/runtime
│   ├── batch.go, cascade.go,...   # CRUD, query execution, schema, secrets internals
│   └── resolver                   # Resolver cache and resolution logic
├── internal                       # Internal-only utilities (HTTP client, schema tools)
├── onyx                           # Public Go SDK surface (init, helpers, condition builders)
├── contract/STABILITY.md          # Contract stability promises
├── onyx.schema.json               # Local schema used for generation/diff
├── onyx-database.json             # Example database config
├── scripts/run-examples.sh        # Helper to run all examples
├── LICENSE                        # License
└── RELEASING.md                   # Release process notes
```
  - API source: `onyx-gen-go --source api`

If you prefer not to install, prefix commands with `go run ./cmd/<cli> ...`.

## Examples

The Go examples mirror the TypeScript examples and build as normal binaries. They rely on `ONYX_DATABASE_ID`, `ONYX_DATABASE_BASE_URL`, `ONYX_DATABASE_API_KEY`, and `ONYX_DATABASE_API_SECRET` (see `examples/shared/config.go` for defaults you can override).

- Compile-check everything: `go test ./examples/...`
- Run a specific sample: `go run ./examples/query/cmd/list` (or any other example under `examples/...`).

## Generated SDK quickstart

Use the generator to emit a typed client into `./gen/onyx`, then initialize and run CRUD.

1) Install the generator (local or global):
```bash
go install ./cmd/onyx-gen-go
# or once published: go install github.com/OnyxDevTools/onyx-database-go/cmd/onyx-gen-go@latest
```

2) Generate the SDK (from schema file or API):
```bash
# From local schema
onyx-gen-go --schema ./onyx.schema.json --out ./gen/onyx --package onyx

# From the API
onyx-gen-go --source api --database-id "$ONYX_DATABASE_ID" --out ./gen/onyx --package onyx
```

3) Initialize the generated client:
```go
package main

import (
	"context"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/gen/onyx"
)

func main() {
	ctx := context.Background()
	db, err := onyx.New(ctx, onyx.Config{
		DatabaseID:      "db_123",
		DatabaseBaseURL: "https://api.onyx.dev",
		APIKey:          "key",
		APISecret:       "secret",
	})
	if err != nil {
		log.Fatal(err)
	}
	// use db.Users(), db.Roles(), etc.
}
```

4) Simple CRUD with the generated client:
```go
// Create
u, err := db.Users().Save(ctx, onyx.User{
	Id:       "user_1",
	Email:    "user@example.com",
	Username: "User One",
})
if err != nil { log.Fatal(err) }

// Read
users, err := db.Users().
	Where(onyx.Contains("email", "@example.com")).
	Limit(10).
	List(ctx)
if err != nil { log.Fatal(err) }

// Update (type-safe updates)
updates := onyx.NewUserUpdates().SetUsername("Updated Name")
count, err := db.Users().
	Where(onyx.Eq("id", u.Id)).
	SetUserUpdates(updates).
	Update(ctx)
if err != nil { log.Fatal(err) }
_ = count // rows affected

// Delete
_, err = db.Users().DeleteByID(ctx, u.Id)
if err != nil { log.Fatal(err) }

// Paging
pages := db.Users().Pages(ctx)
for pages.Next() {
	page, err := pages.Page()
	if err != nil { log.Fatal(err) }
	for _, user := range page.Items {
		log.Println(user.Id)
	}
}
if err := pages.Err(); err != nil { log.Fatal(err) }
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
db, err := onyx.Init(ctx, onyx.Config{/* ... */})
if err != nil { log.Fatal(err) }

page, err := db.Users().
    Select("id", "email", "profile.avatarUrl").
    Where(onyx.Contains("email", "@example.com")).
    And(onyx.Gte("createdAt", "2024-01-01T00:00:00Z")).
    Resolve("roles", "profile").
    OrderBy("createdAt", false).
    Limit(25).
    Page(ctx, "") // empty cursor = first page
if err != nil { log.Fatal(err) }

for _, user := range page.Items {
    fmt.Println(user)
}

iter, err := db.Users().Stream(ctx)
if err != nil { log.Fatal(err) }
for iter.Next() {
    // handle each streamed item
}
if err := iter.Err(); err != nil { log.Fatal(err) }
```

## Resolves and nested data

Use `Resolve` with dotted paths to hydrate related records just like the TS client:
```go
users, err := db.Users().
    Where(onyx.Eq("id", "user_123")).
    Resolve("profile", "roles.permissions").
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
db, err := onyx.Init(ctx, onyx.Config{/* ... */})
if err != nil { log.Fatal(err) }

doc, err := db.Documents().Get(ctx, "doc_123")
if err != nil { log.Fatal(err) }

doc.Data["status"] = "updated"
if _, err := db.Documents().Save(ctx, doc); err != nil {
    log.Fatal(err)
}

secret := onyx.Secret{Key: "API_KEY", Value: "abc"}
if _, err := db.Secrets().Set(ctx, secret); err != nil {
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
onyx-gen-go --schema ./onyx.schema.json --out ./gen/onyx --package onyx --tables User,Order --timestamps time

# Pull the schema from the API and emit the same helpers
onyx-gen-go --source api --database-id db_123 --out ./gen/onyx --package onyx
```

Generated helpers expose typed accessors like `db.Users().Where(onyx.Eq("id", userID)).List(ctx)` while keeping the public API aligned with the TS client.
