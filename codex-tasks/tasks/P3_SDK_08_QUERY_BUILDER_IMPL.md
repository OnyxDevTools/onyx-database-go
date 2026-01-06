---
id: P3_SDK_08_QUERY_BUILDER_IMPL
title: Implement onyx query builder that satisfies contract.Query (build payload only)
depends_on:
- P3_SDK_01_PUBLIC_INIT
- P1_CONTRACT_05_QUERY_INTERFACE
- P1_CONTRACT_04_CONDITIONS
- P1_CONTRACT_03_SORT
tags:
- sdk
- query
---

# Objective

    Implement a concrete `onyx` query builder satisfying `contract.Query`:
    - immutable or copy-on-write chaining
    - stores table, conditions, selects, resolves, sorts, limit


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `onyx/query_builder.go`
    - `onyx/query_payload.go` with a deterministic JSON payload structure
    - tests:
      - chaining correctness
      - deterministic marshaling
      - `Within/NotWithin` subquery embedding works

    ## Acceptance criteria

    - `go test ./...` passes
    - Payload matches TS client structure (verify against TS)
