---
id: P2_SCHEMA_CLI_02_VALIDATE
title: Implement onyx-schema-go validate (local file only)
depends_on:
- P2_SCHEMA_CLI_01_SKELETON
- P1_CONTRACT_07_SCHEMA_TYPES
tags:
- cli
- schema
---

# Objective

    Implement `onyx-schema-go validate` which:
    - reads schema JSON from `--schema` path (default `./onyx.schema.json`)
    - parses via `contract.ParseSchemaJSON`
    - performs semantic validation (at minimum: unique table names, unique field names per table)
    - prints a short success message to stdout on success
    - prints errors to stderr on failure


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `cmd/onyx-schema-go/commands/validate.go`
    - tests:
      - `cmd/onyx-schema-go/commands/validate_test.go` (use temp dirs and sample schemas)

    ## Acceptance criteria

    - Valid schema exits 0
    - Invalid schema exits 1 with useful error message
