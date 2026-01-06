---
id: P2_CODEGEN_06_GOLDEN_TESTS
title: Add golden tests for codegen output stability
depends_on:
- P2_CODEGEN_04_GENERATE_STRUCTS_AND_TABLES
- P2_CODEGEN_05_GENERATE_HELPERS
tags:
- cli
- codegen
- tests
---

# Objective

    Ensure codegen stays stable by adding golden tests.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `cmd/onyx-gen-go/generator/testdata/` with:
      - sample `onyx.schema.json`
      - expected `generated.go.golden`
    - `cmd/onyx-gen-go/generator/generator_test.go` that:
      - runs generator
      - compares to golden

    ## Acceptance criteria

    - Golden tests pass
    - Updating golden requires intentional action
