---
id: P3_SDK_01_PUBLIC_INIT
title: Implement onyx.Init that resolves config and returns contract.Client
depends_on:
- P1_CONTRACT_10_FREEZE_V1
- P0_SETUP_03_CI_BASELINE
tags:
- sdk
- foundation
---

# Objective

    Create the public `onyx` package entrypoints that return a `contract.Client`.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `onyx/client.go`:
      - `type Config struct { ... }` (explicit config)
      - `func Init(ctx context.Context, cfg Config) (contract.Client, error)`
      - `func InitWithDatabaseID(ctx context.Context, databaseID string) (contract.Client, error)`
      - `func ClearConfigCache()`
    - Ensure `onyx` package **implements** `contract.Client` via a concrete type (unexported is fine).

    ## Notes

    - Resolver chain + caching implemented in subsequent tasks.
    - For now, `Init` can call a stub `resolver.Resolve(...)` that you create.

    ## Acceptance criteria

    - `go test ./...` passes
    - `Init` returns a non-nil client in unit tests using stubbed resolver + httptest server
