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

## Initialize the client

The Go client resolves configuration in the same order as the TS client: explicit values > environment variables > config files. Use `ClearConfigCache` between tests to reset cached values.

### Explicit configuration
```go
ctx := context.Background()

client, err := onyx.Init(ctx, onyx.Config{
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
client, err := onyx.InitWithDatabaseID(ctx, "db_123")
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

page, err := client.From("User").
    Select("id", "email", "profile.avatarUrl").
    Where(contract.Contains("email", "@example.com")).
    And(contract.Gte("createdAt", "2024-01-01T00:00:00Z")).
    Resolve("roles", "profile").
    OrderBy(contract.Desc("createdAt")).
    Limit(25).
    Page(ctx, "") // empty cursor = first page
if err != nil { log.Fatal(err) }

for _, user := range page.Items {
    fmt.Println(user)
}

iter, err := client.From("Event").Stream(ctx)
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
post, err := client.From("Post").
    Where(contract.Eq("id", "post_123")).
    Resolve("author.profile", "comments.author").
    Limit(1).
    List(ctx)
if err != nil { log.Fatal(err) }
```

## Cascades

Cascade specs match TS semantics and are constructed via `contract.Cascade`:
```go
spec := contract.Cascade("userRoles:UserRole(userId,id)")

if err := client.Cascade(spec).Save(ctx, "User", user); err != nil {
    log.Fatal(err)
}

if err := client.Cascade(spec).Delete(ctx, "User", user.ID); err != nil {
    log.Fatal(err)
}
```

## Documents and secrets APIs

The Go SDK mirrors the TS helpers for auxiliary APIs as well:

```go
ctx := context.Background()
client := mustInit(ctx)

doc, err := client.Documents().Get(ctx, "doc_123")
if err != nil { log.Fatal(err) }

doc.Data["status"] = "updated"
if _, err := client.Documents().Save(ctx, doc); err != nil {
    log.Fatal(err)
}

secret := contract.Secret{Key: "API_KEY", Value: "abc"}
if _, err := client.Secrets().Set(ctx, secret); err != nil {
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
onyx-schema-go publish --schema ./onyx.schema.json
```

## Codegen CLI (onyx-gen-go)

Generate strongly-typed models, table constants, and helpers with deterministic output that only import `contract` (plus `time` when timestamps are requested):
```bash
onyx-gen-go --schema ./onyx.schema.json --out ./models/onyx.go --package models --tables User,Order --timestamps time

# Pull the schema from the API and emit the same helpers
onyx-gen-go --source api --database-id db_123 --out ./models/onyx.go --package models
```

Generated helpers let you write `q := FromUser(client).Where(contract.Eq("id", userID))` while keeping the public API aligned with the TS client.

## Testing and coverage

CI enforces a coverage threshold. Run the same commands locally to verify:

```bash
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```

## Working with the task pack
- Tasks are dependency-aware: P0 setup → P1 contract freeze → P2 CLIs → P3 SDK core → P4 docs/CI/release/parity.
- Start with `P1_CONTRACT_10_FREEZE_V1` before SDK work; P2/P3 tasks can run in parallel when dependencies allow.
- Use the per-task markdown files in `codex-tasks/tasks/` for acceptance criteria and implementation notes.
