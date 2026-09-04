# Onyx Database Go Client SDK

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![codecov](https://codecov.io/gh/OnyxDevTools/onyx-database-go/branch/main/graph/badge.svg)](https://codecov.io/gh/OnyxDevTools/onyx-database-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/OnyxDevTools/onyx-database-go/onyx.svg)](https://pkg.go.dev/github.com/OnyxDevTools/onyx-database-go/onyx)

Go client SDK for Onyx Cloud Database — a zero-dependency, strict-typed, builder-pattern API for querying and persisting data in Onyx. Includes a credential resolver plus optional schema-driven codegen via the Onyx CLI.

- Website: <https://onyx.dev/>
- Cloud Console: <https://cloud.onyx.dev>
- Docs hub: <https://onyx.dev/documentation/>
- Cloud API docs: <https://onyx.dev/documentation/api-documentation/>
- Examples: `./examples` (separate Go module)

---

## Getting started (Cloud → schema → go generate)

1. **Create a database** at <https://cloud.onyx.dev>. Define your schema (tables like `User`, `Role`, `Permission`) and create API keys. 

2. **Capture connection parameters**:
   You will need to setup an apiKey to connect to your database in the onyx console at <https://cloud.onyx.dev>.  After creating the apiKey, you can download the `onyx-database.json`. Save it to the `config` folder

3. **Install the SDK + CLI**

   Add the SDK to your project (writes to your `go.mod`):
   ```bash
   go get github.com/OnyxDevTools/onyx-database-go@latest
   ```

   Install the Onyx CLI (adds `onyx` to your PATH):
   ```bash
   curl -fsSL https://raw.githubusercontent.com/OnyxDevTools/onyx-cli/main/scripts/install.sh | bash
   ```
   Or via Homebrew:
   ```bash
   brew tap OnyxDevTools/onyx-cli
   brew install onyx-cli
   ```

4. **initialize your generator cofig (go:generate anchor)** :

   ```bash
   onyx init
   ```
   alternativly, you can override the default args like this:

  ```bash
   onyx init --schema ./api/onyx.schema.json --out ./gen/onyx --package onyx
   ```
    
   This creates `generate.go` with the go:generate line and expects this project folder structure:

   ```
   .
   ├── generate.go               # emitted by onyx init
   ├── api/onyx.schema.json      # your schema (download via console or CLI)
   ├── config/onyx-database.json # your onyx connection config, alternatively you can set envars
   └── gen/onyx/                 # generated client lives here
   ```

5. Place your onyx.schema.json file in the api folder of your project.

   > No local schema file? You can fetch it using the schema cli tool `onyx schema get`

6. **Generate the client**:

   ```bash
   go generate
   ```

7. **Start coding!**:

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

       user, err := db.Users().Save(ctx, onyx.User{
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

---

## Client Initialization

This SDK resolves credentials automatically using the chain **explicit config ➜ environment variables ➜ `ONYX_CONFIG_PATH` file ➜ project config file ➜ home profile** _(Node.js only for file-based sources)_. Call `onyx.New(ctx, { DatabaseID: 'database-id' })` to target a specific database, or omit the `databaseId` to use the default. You can also pass credentials directly via config.. Reset caches between tests with `onyx.ClearConfigCache()`.

### Option A) Environment variables

Set credentials, then call `Init` if you use the raw sdk and handle marshalling and decoding yourself, or you can use the generated client which has a `New` method

```bash
export ONYX_DATABASE_ID="db_123"
export ONYX_DATABASE_BASE_URL="https://api.onyx.dev"
export ONYX_DATABASE_API_KEY="key_abc"
export ONYX_DATABASE_API_SECRET="secret_xyz"
```

```go
db, err := onyx.New(ctx, onyx.Config{DatabaseID: "db_123"}) // uses env + cached resolver
if err != nil { log.Fatal(err) }
// Call onyx.ClearConfigCache() when you need to reset cached config between tests
```

### Option B) Explicit config

```go
db, err := onyx.New(ctx, onyx.Config{
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

- `./config/onyx-database-<databaseId>.json`
- `./config/onyx-database.json`
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

`onyx.Init` / `onyx.New` resolve configuration once per cache key and reuse a single signed HTTP client (keep-alive enabled). Reuse the returned client across operations; `CacheTTL` controls how long resolution results are reused. `onyx.ClearConfigCache()` also clears the HTTP client cache.

### Entity transport

Entity saves, deletes, batches, queries, and query streams use the smaller MessagePack transport by default. To force the legacy JSON transport, set `WireFormat` when initializing the core client:

```go
db, err := onyx.Init(ctx, onyx.Config{
    DatabaseID: "db_123",
    WireFormat: onyx.WireFormatJSON,
})
```

The SDK sends `application/vnd.msgpack` and accepts a JSON response as a compatibility fallback. Documents, schemas, secrets, configuration discovery, and AI routes continue to use JSON. MessagePack streams contain concatenated self-delimiting values; the iterator handles framing and any initial `nil` proxy-flush frame.

The binary profile preserves the recursive JSON model: null, booleans, signed 64-bit integers, finite floating-point values, UTF-8 strings, arrays, and string-keyed maps. Binary and extension values are intentionally excluded. Struct `json` tags, `omitempty`, `json.Marshaler`, embedded fields, and `,string` fields keep their existing semantics; values requiring the more complex JSON rules are normalized through `encoding/json` before MessagePack encoding.

Run the native codec and size benchmarks with:

```bash
go test ./internal/msgpack -run '^TestRepresentativeEncodedSize$' -bench 'Benchmark(Encode|Decode)' -benchmem -count=5
```

Reference results from the 250-row fixture (median of five runs, Go 1.22.8, linux/amd64, AMD Ryzen AI MAX+ 395, 2026-08-29):

| Metric | JSON | MessagePack | Difference |
| --- | ---: | ---: | ---: |
| Raw payload | 43,888 bytes | 34,126 bytes | 22.2% smaller |
| Encode | 271,697 ns/op | 124,376 ns/op | 54.2% faster |
| Encode heap | 214,069 B/op, 5,003 allocs/op | 190,161 B/op, 517 allocs/op | 11.2% fewer bytes, 89.7% fewer allocations |
| Decode | 354,771 ns/op | 279,609 ns/op | 21.2% faster |
| Decode heap | 273,293 B/op, 7,978 allocs/op | 282,159 B/op, 15,878 allocs/op | 3.2% more bytes, 99.0% more allocations |

These are codec microbenchmarks, not end-to-end network measurements, and results vary by machine. MessagePack reduced payload size and median runtime for this fixture, while its current generic decoder made more allocations than `encoding/json`.

---

## Optional: generate Go types and table-safe clients

Generate from a file:

```bash
onyx gen --go --schema ./api/onyx.schema.json --out ./gen/onyx --package onyx
```

If you keep the default CLI paths (`./onyx.schema.json`, `./gen/onyx`, package `onyx`), you can also run:
```bash
onyx gen --go
```

Generate from the onyx remote api:

```bash
onyx gen --go --source api --database-id "$ONYX_DATABASE_ID" --out ./gen/onyx --package onyx
```

Scaffold and regenerate with go:generate:

```bash
onyx init #first time setup
go generate
```

Flags:
- `--tables User,Role` to emit a subset
- `--timestamps time|string` to control timestamp field types (`time.Time` vs `string`)

Use the generated client:

```go
import (
	"context"
	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
)
ctx := context.Background()
db, err := onyx.New(ctx, onyx.Config{})

users, err := db.Users().Limit(25).List(ctx)
```

---
## Manage schemas from the CLI

`onyx schema` shares the same credential resolver as the SDK:

```bash
# Inspect resolved config and verify connectivity
onyx schema info # using defaults
onyx schema info --database-id "$ONYX_DATABASE_ID"

# Fetch normalized schema from the API (writes ./api/onyx.schema.json by default)
onyx schema get # using defaults
onyx schema get --out ./api/onyx.schema.json
onyx schema get --tables User,Profile --print   # print subset to stdout

# Validate or normalize a local schema file
onyx schema validate # using defaults
onyx schema validate --schema ./api/onyx.schema.json

# Diff local vs API (or vs another file)
onyx schema diff #using defaults
onyx schema diff --a ./api/onyx.schema.json --b ./next.schema.json
onyx schema diff --a ./api/onyx.schema.json --database-id "$ONYX_DATABASE_ID" --json

# Publish changes (normalize + PUT /schemas/{dbId})
onyx schema publish # using defaults
onyx schema publish --schema ./api/onyx.schema.json --database-id "$ONYX_DATABASE_ID"
```

Omit `--database-id` to rely on env vars or config files like `./config/onyx-database.json` or `~/.onyx/onyx-database.json` (a sample lives at `./examples/config/onyx-database.json`).

---

## AI chat + models (OpenAI-style)

The client also speaks to Onyx AI (OpenAI-compatible, default base `https://ai.onyx.dev`; override with `Config.AIBaseURL` or `ONYX_AI_BASE_URL`). Same API key/secret is reused.

Chat completion (non-streaming):

```go
ctx := context.Background()
db, _ := onyx.Init(ctx, onyx.Config{}) // resolves API key/secret + AI base
resp, err := db.Chat(ctx, onyx.AIChatCompletionRequest{
    Model: "onyx-chat",
    Messages: []onyx.AIChatMessage{
        {Role: "user", Content: "Say hello from Onyx in one short sentence."},
    },
})
if err != nil { log.Fatal(err) }
fmt.Println(resp.Choices[0].Message.Content)
```

List available models:

```go
ctx := context.Background()
db, _ := onyx.Init(ctx, onyx.Config{})
models, err := db.GetModels(ctx)
if err != nil { log.Fatal(err) }
for _, m := range models.Data {
    fmt.Println(m.ID)
}
```

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
onyx.Count
onyx.Sum
onyx.Avg
onyx.Min
onyx.Max
onyx.Median
onyx.Percentile
onyx.Std
onyx.Variance
onyx.Upper
onyx.Lower
onyx.Format
onyx.Substring
onyx.Replace
```

`Distinct()` is a query-builder modifier on `onyx.Query`, for example `db.From("User").Select("email").Distinct().List(ctx)`.

### Lexical, semantic, and hybrid search

Use `SearchWithOptions` when an application needs to choose the search mode.
The query text stays natural language; semantic embedding and hybrid routing
are handled by the database:

```go
results, err := db.From("ActiveDocumentChunk").
    SearchWithOptions(
        "how do i calculate cost per horse",
        onyx.SearchOptions{
            Mode:  onyx.SearchModeHybrid,
            Match: onyx.SearchMatchAny,
        },
    ).
    List(ctx)
if err != nil { log.Fatal(err) }
```

`Mode` accepts `SearchModeLexical`, `SearchModeSemantic`, or
`SearchModeHybrid` and defaults to hybrid. `Match` controls the lexical portion
and defaults to `SearchMatchAny`; use `SearchMatchAll` when every normalized
term is required. `MaxCandidates` defaults to 1000 and may be set from 1
through 5000 (hybrid needs at least 2 so both channels receive a budget).
`MinScore` is optional and, when provided, must be from 0 through 1:

```go
minScore := 0.4
results, err := db.From("ActiveDocumentChunk").
    SearchWithOptions("cost per horse", onyx.SearchOptions{
        Mode:          onyx.SearchModeLexical,
        Match:         onyx.SearchMatchAll,
        MinScore:      &minScore,
        MaxCandidates: 500,
    }).
    List(ctx)
```

Queries using `SearchWithOptions` may be combined with ordinary filters and
are read-only. A query may contain only one high-level search, and it cannot be
combined with another `__full_text__` criterion. Legacy
`Search(queryText, minScore...)` remains unchanged.

To search every eligible unpartitioned searchable table with one global
candidate budget, start from the client instead of a table:

```go
results, err := db.
    SearchWithOptions("how do i calculate cost per horse", onyx.SearchOptions{}).
    Limit(50).
    List(ctx)
```

Semantic and hybrid modes require the database server to configure a
`SearchEmbeddingProvider`. The server uses that provider both when it saves
searchable text and when it embeds query text. Both operations must use the
same model/vector space and stable calibration ID. Rows written before the
provider was enabled (or before its model changed) must be re-saved or handled
by a server-side backfill before semantic search can find them. Lexical mode
does not require an embedding provider. A direct high-level search spans every
concrete partition under one global candidate budget by default; use
`InPartition(...)` to constrain it to one partition. Partitioned tables are not
included in all-table search.

### Advanced vector-managed search

Applications that already produce native semantic signatures or query vectors
can use the lower-level vector-managed contracts. Construct a validated query
and use `SearchVector` on one searchable table:

```go
minScore := 0.42
searchQuery, err := onyx.NewVectorSearchQuery(onyx.VectorSearchQueryInput{
    Text:          "storm warning",
    MinScore:      &minScore,
    MaxCandidates: 500, // optional; default 1000, maximum 5000
})
if err != nil { log.Fatal(err) }

results, err := db.From("Article").SearchVector(searchQuery).List(ctx)
if err != nil { log.Fatal(err) }
```

`VectorSearchQueryInput` also accepts a validated `SemanticVectorSignature`
and the optional `NearbyBucketRadius` and `RequireAllTerms` controls. Semantic
64-bit identifiers and fingerprint words are serialized losslessly as strings.

Physically bounded candidate channels are explicit and read-only.
`SEARCH_CANDIDATES` and `HNSW_CANDIDATES` must be the sole root criterion.
Exactly one ordinary-index `CANDIDATES` condition may instead be combined with
recursively non-negated `AND` predicates. Partitioned tables also require one
concrete `InPartition` value:

```go
// Bounded lexical candidates.
lexical, err := onyx.NewVectorSearchQuery(onyx.VectorSearchQueryInput{
    Text:          "storm warning",
    MaxCandidates: 250,
})
if err != nil { log.Fatal(err) }
rows, err := db.From("Article").ApproximateSearch(lexical).List(ctx)

// Bounded native-HNSW nearest neighbors.
hnsw, err := onyx.NewHNSWSearchQuery(onyx.HNSWSearchQueryInput{
    CalibrationID: -7909761245221418085,
    Vector:        []float64{0.25, -0.5, 0.75},
    MaxCandidates: 100,
    EFSearch:      400,
    MinScore:      &minScore,
})
if err != nil { log.Fatal(err) }
neighbors, err := db.From("Article").HNSWCandidates(hnsw).List(ctx)

// Bounded EQUAL/IN admission from an ordinary secondary index.
sample, err := db.From("Article").
    InPartition("tenant-id").
    Where(onyx.ApproximateCandidates(
        "corpusId",
        []string{"public", "archive"},
        100,
    )).
    And(onyx.Eq("published", true)).
    List(ctx)
```

The `CANDIDATES` index walk admits at most `MaxCandidates` rows once; later
`AND` conditions filter only that admitted set. `OR`, negation, and a second
`CANDIDATES` condition are rejected. `Query.ApproximateCandidates` remains as
a deprecated compatibility shortcut; new code should use the condition helper
inside `Where` as shown above.

HNSW vectors contain 1–16384 finite values with a non-zero norm.
`MaxCandidates` is bounded to 1–5000, `EFSearch` defaults to
`max(1000, MaxCandidates)` and is bounded through 20000, and HNSW `MinScore`
is optional in `[-1, 1]`.

### Inner queries (IN/NOT IN)

```go
db, _ := onyx.New(ctx, onyx.Config{})
core := db.Core()

admins, _ := db.Users().
    Where(onyx.Within(
        "id",
        core.From(onyx.Tables.UserRole).
            Select("userId").
            Where(onyx.Eq("roleId", "role-admin")),
    )).
    List(ctx)

rolesMissingPerm, _ := db.Roles().
    Where(onyx.NotWithin(
        "id",
        core.From(onyx.Tables.RolePermission).
            Select("roleId").
            Where(onyx.Eq("permissionId", "perm-manage-users")),
    )).
    List(ctx)
```

`Within`/`NotWithin` accept another query; the SDK serializes the inner query before sending it to the API.

---

### Aggregate + group-by

```go
import coreonyx "github.com/OnyxDevTools/onyx-database-go/onyx"

db, _ := coreonyx.Init(ctx, coreonyx.Config{})

rows, err := db.From("User").
    Select("isActive", coreonyx.Count("id")).
    GroupBy("isActive").
    List(ctx)
if err != nil { log.Fatal(err) }

for _, row := range rows {
    fmt.Printf("isActive=%v count=%v\n", row["isActive"], row["count(id)"])
}
```

### Time buckets with `Format`

```go
import coreonyx "github.com/OnyxDevTools/onyx-database-go/onyx"

db, _ := coreonyx.Init(ctx, coreonyx.Config{})

bucket := coreonyx.Format("dateTime", "yyyy-MM-dd HH")
rows, err := db.From("AuditLog").
    Select(bucket, coreonyx.Count("*"), "status").
    GroupBy(bucket, "status").
    List(ctx)
if err != nil { log.Fatal(err) }

for _, row := range rows {
    fmt.Printf("%v %v %v\n", row[bucket], row["status"], row["count(*)"])
}
```

Aliases are not supported yet, so aggregate keys remain the raw function expressions such as `count(id)` or `format(dateTime,"yyyy-MM-dd HH")`.

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
_, err := db.Users().Save(ctx, onyx.User{
    Id:       "user_124",
    Email:    "bob@example.com",
    Username: "Bob",
})
if err != nil { log.Fatal(err) }

// Batch upsert with typed helper
_, err = db.Users().SaveMany(ctx, []onyx.User{
    {Id: "user_125", Email: "carol@example.com", Username: "Carol"},
    {Id: "user_126", Email: "dana@example.com", Username: "Dana"},
})
if err != nil { log.Fatal(err) }

// Cascade save relationships (uses resolver graph)
cascade := onyx.Cascade("userRoles:UserRole(userId,id)")
_, err = db.Users().Save(ctx, onyx.User{
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
updates := onyx.NewUserUpdates().
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

Searchable table and native index metadata are preserved during fetch,
normalize, validate, and publish round trips:

```go
searchable := onyx.Table{
    Name:          "Article",
    Type:          onyx.TableTypeSearchable,
    SearchSupport: onyx.SearchSupportBoth,
    Indexes: []onyx.Index{
        {Name: "content", Type: onyx.IndexTypeVector},
    },
}
```

Choose `SearchSupportLexical`, `SearchSupportSemantic`, or `SearchSupportBoth`.
Omitting `SearchSupport` retains the backward-compatible `BOTH` behavior for a
searchable table. The field is ignored for ordinary tables.

### Secrets API

```go
secrets := db.OnyxSecrets()
if _, err := secrets.Set(ctx, onyx.OnyxSecret{Key: "api-key", Value: "super-secret"}); err != nil {
    log.Fatal(err)
}
secret, err := secrets.Get(ctx, "api-key")
if err != nil { log.Fatal(err) }
fmt.Println(secret.Value)
if err := secrets.Delete(ctx, "api-key"); err != nil {
    log.Fatal(err)
}
```

### Documents API

```go
doc := onyx.OnyxDocument{
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
