---
id: P1_CONTRACT_03_SORT
title: Define contract.Sort with Asc/Desc helpers (pure JSON shapes)
depends_on:
- P1_CONTRACT_01_PACKAGE_SKELETON
tags:
- contract
- query
---

# Objective

    Define sorting primitives in `/contract`:
    - `Sort` interface with `MarshalJSON()`
    - helpers `Asc(field)` and `Desc(field)` that return a concrete type


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `contract/sort.go`
    - `contract/sort_test.go` with golden JSON assertions

    ## Acceptance criteria

    - `go test ./contract -run TestSort` passes
    - JSON output is stable and minimal
