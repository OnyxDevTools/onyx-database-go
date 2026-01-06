---
id: P1_CONTRACT_08_CLIENT_INTERFACE
title: Define contract.Client interface (From, Cascade, Save/Delete/BatchSave/Schema)
depends_on:
- P1_CONTRACT_05_QUERY_INTERFACE
- P1_CONTRACT_06_CASCADE
- P1_CONTRACT_07_SCHEMA_TYPES
- P1_CONTRACT_02_ERRORS
tags:
- contract
- client
---

# Objective

    Create the stable `contract.Client` interface for the Go SDK.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `contract/client.go` with:
      - `From(table string) Query`
      - `Cascade(spec CascadeSpec) CascadeClient`
      - `Save(ctx, table string, entity any) error`
      - `Delete(ctx, table, id string) error`
      - `BatchSave(ctx, table string, entities []any, batchSize int) error`
      - `Schema(ctx) (Schema, error)`

    ## Acceptance criteria

    - `go test ./contract` passes
    - No non-stdlib deps
