---
id: P2_CODEGEN_02_LOAD_SCHEMA
title: Implement schema loader + normalization for codegen
depends_on:
- P2_CODEGEN_01_SKELETON
- P1_CONTRACT_07_SCHEMA_TYPES
tags:
- cli
- codegen
---

# Objective

    Implement schema loading pipeline used by codegen:
    - read JSON file
    - parse with `contract.ParseSchemaJSON`
    - normalize with `contract.NormalizeSchema`
    - filter tables by `--tables` if set


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `cmd/onyx-gen-go/generator/load.go`
    - unit tests using sample schema JSONs

    ## Acceptance criteria

    - Deterministic table/field order in the model passed to templates
