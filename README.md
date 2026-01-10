# Onyx Database Go SDK

Go SDK and CLIs for Onyx Database, with a contract-first design based on generating a client based on the onyx schema file. 
Also contains 

## Prereqs

Requires Go 1.22+. Install the CLIs into your `$(go env GOBIN)` (or `$(go env GOPATH)/bin`) from the repo root:

## Getting Started

A step-by-step flow from a fresh checkout:

1) Install the unified CLI:
```bash
go install github.com/OnyxDevTools/onyx-database-go/cmd/onyx-go@latest
```

2) Scaffold a go:generate anchor (run from your project root):
```bash
onyx-go gen init --schema ./api/onyx.schema.json --out ./gen/onyx --package onyx
```
This assumes:
```
.
├── generate.go           # created by the command above
├── api/onyx.schema.json  # your schema
└── gen/onyx/             # will contain generated code after go generate
```

3) Generate the client:
```bash
go generate   # runs the go:generate line emitted in generate.go
```

4) Initialize once at startup and reuse:
```go
package main

import (
    "context"
    "log"

    onyx "your/module/gen/onyx"
)

func main() {
    ctx := context.Background()
    db, err := onyx.New(ctx, onyx.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Save
    user, err := db.Users().Save(ctx, onyx.User{
        Id:       "user_1",
        Email:    "user@example.com",
        Username: "User One",
    })
    if err != nil { log.Fatal(err) }
    _ = user
}
```

5) (Optional) If you dont have a local schema file but have a database setup in cloud.onyx.dev, you can retreive it:
```bash
onyx-go schema get
```

5) (Optional) Regenerate from the API instead of a file:
```bash
onyx-go gen --source api --database-id "$ONYX_DATABASE_ID" --out ./gen/onyx --package onyx
```


## CLI tools 
- Schema CLI:
  - Validate: `onyx-schema-go validate`
  - Diff: `onyx-schema-go diff`
  - Get (API): `onyx-schema-go get` (replaces your schema with what is in onyx) or use `--print` to pretty-print without writing to your local schema)
  - Publish (API): `onyx-schema-go publish` if valid, will publish your remote database with the local one

- Codegen CLI:
  - Example regen (in this repo): `go generate ./...` (writes to `./examples/gen/onyx`)

## Client Initalization

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

You can also supply a JSON file (lowest precedence) named `onyx-database.json`, `onyx-database-<dbid>.json`, or pointed to via `ONYX_CONFIG_PATH` (a sample lives at `./examples/config/onyx-database.json`):
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
onyx-schema-go validate --schema ./examples/api/onyx.schema.json
onyx-schema-go diff --a ./examples/api/onyx.schema.json --b ./next.schema.json --json
onyx-schema-go normalize --schema ./examples/api/onyx.schema.json --out ./examples/api/onyx.normalized.json
onyx-schema-go get --database-id db_123 --out ./examples/api/onyx.schema.json
# or just print the remote schema using your env/config (e.g., onyx-database.json; sample at ./examples/config/onyx-database.json):
onyx-schema-go get --print
onyx-schema-go publish --schema ./examples/api/onyx.schema.json
```

If `--database-id` is omitted, the CLI resolves credentials from env vars or config files such as `./onyx-database.json` (also `~/.onyx/onyx-database.json`; this repo includes `./examples/config/onyx-database.json` as a sample).

## Codegen CLI (onyx-gen-go)

Generate strongly-typed models, table constants, and helpers with deterministic output that only import `contract` (plus `time` when timestamps are requested):
```bash
# In this repo (writes to examples/gen/onyx with a fixed timestamp)
go generate ./...

# Manual regeneration against local schema (same as go generate)
onyx-gen-go --schema ./examples/api/onyx.schema.json --out ./examples/gen/onyx --package onyx

# Pull the schema from the API and emit the same helpers (examples module)
onyx-gen-go --source api --database-id db_123 --out ./examples/gen/onyx --package onyx
```

Generated helpers expose typed accessors like `db.Users().Where(onyx.Eq("id", userID)).List(ctx)` while keeping the public API aligned with the TS client. In your own app, set `--out` to your repository path and import that package (do not import the example package from this repo; the copy under `examples/gen/onyx` exists only for the samples in this repo).
