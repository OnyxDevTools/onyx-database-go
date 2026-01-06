---
id: P2_CODEGEN_04_GENERATE_STRUCTS_AND_TABLES
title: Generate Go structs + table constants (imports contract only if helpers emitted)
depends_on:
- P2_CODEGEN_03_TYPE_MAPPING
tags:
- cli
- codegen
---

# Objective

    Generate:
    - per-table Go structs with JSON tags
    - `const` table names like `TableUser = "User"`
    - no timestamps in file header
    - stable output formatting (`gofmt`)


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `cmd/onyx-gen-go/generator/render.go` using `text/template` or `go/format`
    - templates under `cmd/onyx-gen-go/templates/`
    - ensure generated file compiles

    ## Acceptance criteria

    - generated output is deterministic across runs
    - `gofmt`-formatted
