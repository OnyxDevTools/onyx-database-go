---
id: P2_CODEGEN_05_GENERATE_HELPERS
title: Generate optional helper functions (FromUser etc) that depend only on contract
depends_on:
- P2_CODEGEN_04_GENERATE_STRUCTS_AND_TABLES
tags:
- cli
- codegen
---

# Objective

    Add optional helper emission (enabled by default):
    - `func FromUser(c contract.Client) contract.Query { return c.From(TableUser) }`


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Requirements

    - Generated code must import:
      - `github.com/OnyxDevTools/onyx-database-go/contract`
      - plus `time` if needed
    - It must NOT import `onyx` or any internal packages.

    ## Required deliverables

    - Template updates + tests ensuring imports are minimal and correct.

    ## Acceptance criteria

    - `go test ./...` passes after generating into a temp module in tests OR compiling generated snippets.
