# onyx-database-go

Go SDK and CLIs for Onyx Database, built with a contract-first design that mirrors the TypeScript client. The repository is currently scaffolded with the core packages and CLI entry points that future tasks will fill in.

## Status and layout
- `contract/` — stable, stdlib-only API surface guarded by compliance tests and `STABILITY.md`.
- `onyx/` — SDK implementation: configuration resolution, HTTP client, auth/signing, query execution, CRUD/cascade/schema, documents and secrets endpoints.
- `internal/` — shared helpers kept out of the public API.
- `cmd/onyx-schema-go/` — schema CLI: validate, diff, normalize, get, publish.
- `cmd/onyx-gen-go/` — codegen CLI: Go structs/table consts/helpers from `onyx.schema.json`.
- `examples/` — runnable samples once the SDK is implemented.
- `codex-tasks/` — task manifest (`tasks.yaml`) plus per-task guides.

## Development principles
- Contract-first: `/contract` stays stdlib-only and stable; breaking changes require a deliberate major bump and updates to `contract/STABILITY.md`.
- Determinism: stable JSON for conditions/sorts, normalized schema ordering, reproducible codegen with no timestamps; generated files import `contract` only (plus `time` when needed).
- Testing: every behavior in `/onyx`, `/cmd`, `/internal` needs unit tests. Run `go test ./...` and `go vet ./...` (optionally `golangci-lint run`).
- Resolver parity: explicit config > env vars > config files with cache TTL + `ClearConfigCache`, matching the TS client.

## Planned usage (post-implementation)

### Initialize the client
```go
ctx := context.Background()

client, err := onyx.Init(ctx, onyx.Config{
    DatabaseID:      "db_123",
    DatabaseBaseURL: "https://api.onyx.dev",
    APIKey:          os.Getenv("ONYX_DATABASE_API_KEY"),
    APISecret:       os.Getenv("ONYX_DATABASE_API_SECRET"),
})
if err != nil { log.Fatal(err) }

// Or rely on env/config files:
client, err = onyx.InitWithDatabaseID(ctx, "db_123")
```

### Query data
```go
users, err := client.From("User").
    Where(contract.Contains("email", "@example.com")).
    Resolve("roles").
    OrderBy(contract.Desc("createdAt")).
    Limit(50).
    List(ctx)
if err != nil { log.Fatal(err) }

iter, err := client.From("Event").Stream(ctx)
for iter.Next() {
    evt := iter.Value()
    // handle event
}
if err := iter.Err(); err != nil { log.Fatal(err) }
```

### Cascade saves
```go
spec := contract.Cascade("userRoles:UserRole(userId,id)")
if err := client.Cascade(spec).Save(ctx, "User", user); err != nil {
    log.Fatal(err)
}
```

### Schema CLI (onyx-schema-go)
```
onyx-schema-go validate --schema ./onyx.schema.json
onyx-schema-go diff --a ./onyx.schema.json --b ./next.schema.json --json
onyx-schema-go normalize --schema ./onyx.schema.json --out ./onyx.normalized.json
onyx-schema-go get --database-id db_123 --out ./onyx.schema.json
onyx-schema-go publish --schema ./onyx.schema.json
```

### Codegen CLI (onyx-gen-go)
```
onyx-gen-go --schema ./onyx.schema.json --out ./models/onyx.go --package models --tables User,Order --timestamps time
onyx-gen-go --source api --database-id db_123 --out ./models/onyx.go --package models
```
Generated helpers should let you write `from := FromUser(client)` and work with `TableUser` constants while importing only `contract`.

## Working with the task pack
- Tasks are dependency-aware: P0 setup → P1 contract freeze → P2 CLIs → P3 SDK core → P4 docs/CI/release/parity.
- Start with `P1_CONTRACT_10_FREEZE_V1` before SDK work; P2/P3 tasks can run in parallel when dependencies allow.
- Use the per-task markdown files in `codex-tasks/tasks/` for acceptance criteria and implementation notes.
