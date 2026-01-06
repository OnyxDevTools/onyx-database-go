---
id: P2_SCHEMA_CLI_03_DIFF
title: Implement onyx-schema-go diff (local vs local) with --json option
depends_on:
- P2_SCHEMA_CLI_01_SKELETON
- P1_CONTRACT_07_SCHEMA_TYPES
tags:
- cli
- schema
---

# Objective

    Implement `onyx-schema-go diff` which compares two local schema files:
    - `--a` path (default `./onyx.schema.json`)
    - `--b` path (required)

    Output:
    - default: human-readable summary to stdout
    - with `--json`: machine-readable JSON diff


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Minimum diff model

    - Added/removed tables
    - Added/removed fields per table
    - Changed field type/nullable

    ## Required deliverables

    - `cmd/onyx-schema-go/commands/diff.go`
    - `internal/schema/diff.go` (implementation helper is fine)
    - unit tests

    ## Acceptance criteria

    - Deterministic diff ordering
    - `--json` output stable for tests (golden)
