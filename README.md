# Onyx Database Go SDK

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![codecov](https://codecov.io/gh/OnyxDevTools/onyx-database-go/branch/main/graph/badge.svg)](https://codecov.io/gh/OnyxDevTools/onyx-database-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/OnyxDevTools/onyx-database-go/onyx.svg)](https://pkg.go.dev/github.com/OnyxDevTools/onyx-database-go/onyx)

Go SDK and CLIs for **Onyx Cloud Database** with a contract-first design and stable codegen. The SDK is stdlib-only, resolver/query surface, supports streaming, secrets, documents, and ships a generator for table-safe Go types.

- Website: <https://onyx.dev/>
- Cloud Console: <https://cloud.onyx.dev>
- Docs hub: <https://onyx.dev/documentation/>
- Cloud API docs: <https://onyx.dev/documentation/api-documentation/>
- Examples: `./examples` (separate Go module)

---

## Getting started (Cloud → schema → go generate)

1. **Create a database** at <https://cloud.onyx.dev>. Define your schema (tables like `User`, `Role`, `Permission`) and create API keys.
2. **Capture connection parameters**:
   - `baseUrl` (e.g., `https://api.onyx.dev`)
   - `databaseId`
   - `apiKey`
   - `apiSecret`
3. **Install the CLI tool and sdk libraryr (Go 1.22+):

   ```bash
   go install github.com/OnyxDevTools/onyx-database-go/cmd/onyx-go@latest
   go install github.com/OnyxDevTools/onyx-database-go/onyx@latest
   ```

4. **Scaffold a go:generate anchor** (run in your project):

   ```bash
   onyx-go gen init --schema ./onyx.schema.json --out ./gen/onyx --package onyx
   ```

   This creates `generate.go` with the go:generate line and expects:

   ```
   .
   ├── generate.go          # emitted by onyx-go gen init
   ├── onyx.schema.json     # your schema (download via console or CLI)
   └── gen/onyx/            # generated client lives here
   ```

5. **Generate the client**:

   ```bash
   go generate ./...
   ```

6. **Code usage example **:

   ```go
   package main

   import (
       "context"
       "log"

       client "your/module/gen/onyx"
   )

   func main() {
       ctx := context.Background()
       db, err := client.New(ctx, client.Config{})
       if err != nil {
           log.Fatal(err)
       }

       user, err := db.Users().Save(ctx, client.User{
           Id:       "user_1",
           Email:    "user@example.com",
           Username: "User One",
       })
       if err != nil {
           log.Fatal(err)
       }
       _ = user
   }
   ```

> No local schema file? Fetch it from the API with `onyx-go schema get --out ./onyx.schema.json`, then regenerate with `onyx-go gen --source api ...`.

---

## Install

- Add the SDK to your module (the generator and examples depend on it):

  ```bash
  go get github.com/OnyxDevTools/onyx-database-go/onyx
  ```

- Install CLIs into `$(go env GOBIN)` (or `$(go env GOPATH)/bin`):

  ```bash
  go install github.com/OnyxDevTools/onyx-database-go/cmd/onyx-go@latest
  ```

The runtime has no external dependencies beyond the Go stdlib. Works anywhere Go 1.22+ can run; reuse a single client per process for connection pooling.

---

## Initialize the client

Configuration resolution matches the TypeScript SDK: **explicit config → environment variables → config files**, cached for 5 minutes by default. Reset caches between tests with `onyx.ClearConfigCache()`.

`Config` fields: `DatabaseID`, `DatabaseBaseURL`, `APIKey`, `APISecret`, `CacheTTL`, `ConfigPath`, `LogRequests`, `LogResponses`, and optional `HTTPClient`, `Clock`, `Sleep` overrides for custom transport/testing. Setting `ONYX_DEBUG=true` forces request/response logging even if the flags are false.

### Option A) Environment variables

Set credentials, then call `Init` (or the generated `New`) with an empty config or just the database ID:

```bash
export ONYX_DATABASE_ID="db_123"
export ONYX_DATABASE_BASE_URL="https://api.onyx.dev"
export ONYX_DATABASE_API_KEY="key_abc"
export ONYX_DATABASE_API_SECRET="secret_xyz"
```

```go
db, err := client.New(ctx, client.Config{DatabaseID: "db_123"}) // uses env + cached resolver
if err != nil { log.Fatal(err) }
// Call onyx.ClearConfigCache() when you need to reset cached config between tests
```

### Option B) Explicit config

```go
db, err := client.New(ctx, client.Config{
    DatabaseID:      "db_123",
    DatabaseBaseURL: "https://api.onyx.dev",
    APIKey:          os.Getenv("ONYX_DATABASE_API_KEY"),
    APISecret:       os.Getenv("ONYX_DATABASE_API_SECRET"),
    CacheTTL:        10 * time.Minute, // optional; defaults to 5m
    LogRequests:     true,             // optional request logging
    LogResponses:    false,            // optional response logging
})
if err != nil { log.Fatal(err) }
```

### Option C) Config files (Go-only)

`ConfigPath` or `ONYX_CONFIG_PATH` can point to a JSON file. When unset, the resolver checks (in order):

- `./onyx-database-<databaseId>.json`
- `./onyx-database.json`
- `~/.onyx/onyx-database-<databaseId>.json`
- `~/.onyx/onyx-database.json`
- `~/onyx-database.json`

Shape:

```json
{
  "databaseId": "db_123",
  "databaseBaseUrl": "https://api.onyx.dev",
  "apiKey": "key_abc",
  "apiSecret": "secret_xyz"
}
```

### Connection handling

`onyx.Init` / `client.New` resolve configuration once per cache key and reuse a single signed HTTP client (keep-alive enabled). Reuse the returned client across operations; `CacheTTL` controls how long resolution results are reused. `onyx.ClearConfigCache()` also clears the HTTP client cache.

---

## Optional: generate Go types and table-safe clients

`onyx-go gen` emits:
- Plain Go structs for each table, JSON-tagged to match the API.
- `Tables` constants and `Resolvers` map (when resolvers exist).
- A typed `DB` wrapper with table-specific clients (`Users()`, `Roles()`, etc.), `Documents()`, `Secrets()`, `Core()`, and `Wrap(core)` for adapters.
- Typed helpers: `FindByID`, `FindByEmail` (when an `email` field exists), `FindActiveUsers`/`CountActive` (when an `isActive` field exists), `SaveMany`, `DeleteByIDs`, paginated iterators, `WithTimeout`, `QueryHook`, and cascade support via `...onyx.CascadeSpec`.

Stable ordering (set `ONYX_GEN_TIMESTAMP` to pin header text); imports only `github.com/OnyxDevTools/onyx-database-go/onyx` plus `time` when timestamps are time values.

Generate from a file:

```bash
onyx-go gen
```

is the same as running these default switches: 
```bash
onyx-go gen --schema ./onyx.schema.json --out ./gen/onyx --package onyx
```

Generate from the onyx remote api:

```bash
onyx-go gen --source api --database-id "$ONYX_DATABASE_ID" --out ./gen/onyx --package onyx
```

Scaffold and regenerate with go:generate:

```bash
onyx-go gen init #first time setup
go generate
```

Flags:
- `--tables User,Role` to emit a subset
- `--timestamps time|string` to control timestamp field types (`time.Time` vs `string`)

Use the generated client:

```go
db, err := client.New(ctx, client.Config{})

users, err := db.Users().
    Resolve("roles.permissions").
    OrderBy("createdAt", true).
    Limit(25).
    List(ctx)
```

---
## Manage schemas from the CLI

`onyx-go schema` shares the same credential resolver as the SDK:

```bash
# Inspect resolved config and verify connectivity
onyx-go schema info # using defaults
onyx-go schema info --database-id "$ONYX_DATABASE_ID"

# Fetch normalized schema from the API (writes ./onyx.schema.json by default)
onyx-go schema get. # using defaults
onyx-go schema get --out ./onyx.schema.json
onyx-go schema get --tables User,Profile --print   # print subset to stdout

# Validate or normalize a local schema file
onyx-go schema validate # using defaults
onyx-go schema validate --schema ./onyx.schema.json

# Diff local vs API (or vs another file)
onyx-go schema diff #using defaults
onyx-go schema diff --a ./onyx.schema.json --b ./next.schema.json
onyx-go schema diff --a ./onyx.schema.json --database-id "$ONYX_DATABASE_ID" --json

# Publish changes (normalize + PUT /schemas/{dbId})
onyx-go schema publish # using defaults
onyx-go schema publish --schema ./onyx.schema.json --database-id "$ONYX_DATABASE_ID"
```

Omit `--database-id` to rely on env vars or config files like `./onyx-database.json` or `~/.onyx/onyx-database.json` (a sample lives at `./examples/config/onyx-database.json`).

---

## Query helpers at a glance

```go
import (
    "github.com/OnyxDevTools/onyx-database-go/onyx"
)

onyx.Eq
onyx.Neq
onyx.In
onyx.NotIn
onyx.Between
onyx.Gt
onyx.Gte
onyx.Lt
onyx.Lte
onyx.Like
onyx.Contains
onyx.StartsWith
onyx.IsNull
onyx.NotNull
onyx.Within      // IN subquery
onyx.NotWithin   // NOT IN subquery
onyx.Asc
onyx.Desc
```

### Inner queries (IN/NOT IN)

```go
db, _ := client.New(ctx, client.Config{})
core := db.Core()

admins, _ := db.Users().
    Where(onyx.Within(
        "id",
        core.From(client.Tables.UserRole).
            Select("userId").
            Where(onyx.Eq("roleId", "role-admin")),
    )).
    List(ctx)

rolesMissingPerm, _ := db.Roles().
    Where(onyx.NotWithin(
        "id",
        core.From(client.Tables.RolePermission).
            Select("roleId").
            Where(onyx.Eq("permissionId", "perm-manage-users")),
    )).
    List(ctx)
```

`Within`/`NotWithin` accept another query; the SDK serializes the inner query before sending it to the API.

---

## Usage examples (User / Role / Permission)

> Replace `client` with your generated package import (default package name is `onyx`).

### List & page

```go
page, err := db.Users().
    Where(onyx.Eq("isActive", true)).
    And(onyx.Contains("email", "@example.com")).
    Resolve("roles.permissions", "profile").
    OrderBy("createdAt", true).
    Limit(25).
    Page(ctx, "")
if err != nil { log.Fatal(err) }
for _, u := range page.Items {
    fmt.Println(u.Email)
}

// Iterate all pages
iter := db.Users().Pages(ctx)
for iter.Next() {
    p, _ := iter.Page()
    for _, u := range p.Items {
        fmt.Println(u.Id)
    }
}
if err := iter.Err(); err != nil { log.Fatal(err) }
```

### Save / upsert (single, batch, cascade)

```go
// Single upsert
_, err := db.Users().Save(ctx, client.User{
    Id:       "user_124",
    Email:    "bob@example.com",
    Username: "Bob",
})
if err != nil { log.Fatal(err) }

// Batch upsert with typed helper
_, err = db.Users().SaveMany(ctx, []client.User{
    {Id: "user_125", Email: "carol@example.com", Username: "Carol"},
    {Id: "user_126", Email: "dana@example.com", Username: "Dana"},
})
if err != nil { log.Fatal(err) }

// Cascade save relationships (uses resolver graph)
cascade := onyx.Cascade("userRoles:UserRole(userId,id)")
_, err = db.Users().Save(ctx, client.User{
    Id:       "user_200",
    Email:    "cathy@example.com",
    Username: "Cathy",
    UserRoles: []any{
        map[string]any{"roleId": "role_admin"},
        map[string]any{"roleId": "role_editor"},
    },
}, cascade)
if err != nil { log.Fatal(err) }

// Core client batch save (arrays of maps/structs), default chunk size 500
core := db.Core()
_ = core.BatchSave(ctx, "User", []any{{"id": "user_300", "email": "eve@example.com"}}, 0)
```

### Delete (by id or by query)

```go
// Primary-key delete
deleted, err := db.Users().DeleteByID(ctx, "user_125")
if err != nil { log.Fatal(err) }
fmt.Println("rows removed:", deleted)

// Delete matching a query
count, err := db.Users().
    Where(onyx.Eq("isActive", false)).
    Delete(ctx)
if err != nil { log.Fatal(err) }
fmt.Println("inactive removed:", count)
```

### Update in place

```go
now := time.Now().UTC()
updates := client.NewUserUpdates().
    SetLastLoginAt(&now).
    SetIsActive(true)

modified, err := db.Users().
    Where(onyx.Eq("email", "alice@example.com")).
    SetUserUpdates(updates).
    Update(ctx)
if err != nil { log.Fatal(err) }
fmt.Println("rows updated:", modified)
```

### Schema API

```go
core := db.Core()
schema, _ := core.Schema(ctx)
history, _ := core.GetSchemaHistory(ctx)
_ = core.UpdateSchema(ctx, schema, true) // publish=true
fmt.Println("tables:", len(schema.Tables), "history entries:", len(history))
```

### Secrets API

```go
secClient := db.Secrets()
_, _ = secClient.Set(ctx, onyx.Secret{Key: "api-key", Value: "super-secret"})
secret, _ := secClient.Get(ctx, "api-key")
fmt.Println(secret.Value)
_ = secClient.Delete(ctx, "api-key")
```

### Documents API

```go
doc := onyx.Document{
    DocumentID: "logo.png",
    Path:       "/brand/logo.png",
    MimeType:   "image/png",
    Content:    base64LogoPNG,
}
saved, _ := db.Documents().Save(ctx, doc)
fetched, _ := db.Documents().Get(ctx, saved.DocumentID)
_ = db.Documents().Delete(ctx, fetched.DocumentID)
```

### Streaming

```go
iter, err := db.Users().
    Where(onyx.Eq("status", "active")).
    Stream(ctx)
if err != nil { log.Fatal(err) }
defer iter.Close()

for iter.Next() {
    fmt.Println("event:", iter.Value())
}
if err := iter.Err(); err != nil { log.Fatal(err) }
```

---

## Error handling

SDK errors use `*onyx.Error` (code, message, `Meta` with HTTP status, etc.). Use `errors.As` to inspect:

```go
if err != nil {
    var oe *onyx.Error
    if errors.As(err, &oe) {
        fmt.Println("code:", oe.Code, "status:", oe.Meta["status"])
    }
}
```

Initialization failures raise configuration errors; HTTP calls surface server messages and statuses via the same type.

---

## Examples

`./examples` is a standalone Go module with ready-to-run samples for queries, cascades, streaming, schema/diff/publish, documents, and secrets. Point it at your database by setting the same env vars or config file described above.

---

## Release workflow

See `RELEASING.md` for details. In short:

1. `go vet ./...`
2. `go test ./... -coverprofile=coverage.out -covermode=atomic`
3. Update docs/examples as needed.
4. Tag and push `vX.Y.Z`.

---

## License

MIT © Onyx Dev Tools. See `LICENSE`.
