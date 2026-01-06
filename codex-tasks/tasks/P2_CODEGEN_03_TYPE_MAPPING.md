---
id: P2_CODEGEN_03_TYPE_MAPPING
title: Implement Onyx->Go type mapping rules for fields
depends_on:
- P2_CODEGEN_02_LOAD_SCHEMA
tags:
- cli
- codegen
---

# Objective

    Define deterministic mapping from schema field types -> Go types.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Requirements

    - Provide a single function:
      `func GoTypeForField(fieldType string, nullable bool, timestampsMode string) (goType string, needsPointer bool)`
    - Handle at least:
      - string, int, float, bool
      - json / object (map[string]any)
      - arrays (slices)
      - id fields (string)
      - timestamps (time.Time or string via flag)
    - Nullability strategy:
      - for scalar types: pointers when nullable
      - for reference types: `map/slice` already nil-able, but be consistent

    ## Required deliverables

    - `cmd/onyx-gen-go/generator/types.go`
    - unit tests for mapping matrix

    ## Acceptance criteria

    - Mapping rules are documented and tested
