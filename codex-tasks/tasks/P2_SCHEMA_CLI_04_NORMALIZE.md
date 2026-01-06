---
id: P2_SCHEMA_CLI_04_NORMALIZE
title: Implement onyx-schema-go normalize command (optional but recommended)
depends_on:
- P2_SCHEMA_CLI_01_SKELETON
- P1_CONTRACT_07_SCHEMA_TYPES
tags:
- cli
- schema
---

# Objective

    Add `onyx-schema-go normalize` to:
    - read schema JSON
    - normalize ordering via `contract.NormalizeSchema`
    - write normalized JSON to stdout or `--out` file

    This makes diffs deterministic and helps codegen.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `cmd/onyx-schema-go/commands/normalize.go`
    - tests for deterministic output

    ## Acceptance criteria

    - Normalizing twice yields identical bytes
