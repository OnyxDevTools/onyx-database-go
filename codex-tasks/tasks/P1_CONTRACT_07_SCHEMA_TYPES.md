---
id: P1_CONTRACT_07_SCHEMA_TYPES
title: Define schema contract types and JSON parsing helpers
depends_on:
- P1_CONTRACT_01_PACKAGE_SKELETON
tags:
- contract
- schema
---

# Objective

    Define minimal schema types in `/contract` and helpers to parse `onyx.schema.json`.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `contract/schema.go`:
      - `Schema`, `Table`, `Field` interfaces OR concrete structs (choose what keeps contract simplest).
      - Must support:
        - enumerating tables and fields
        - table lookup by name
    - `contract/schema_json.go`:
      - `func ParseSchemaJSON(data []byte) (Schema, error)`
      - `func NormalizeSchema(s Schema) Schema` (stable ordering)
    - `contract/schema_test.go`:
      - parse + normalize round-trip tests

    ## Acceptance criteria

    - `go test ./contract -run TestSchema` passes
    - No external deps
