---
id: P1_CONTRACT_05_QUERY_INTERFACE
title: Define contract.Query interface (builder + terminal ops)
depends_on:
- P1_CONTRACT_01_PACKAGE_SKELETON
- P1_CONTRACT_03_SORT
- P1_CONTRACT_04_CONDITIONS
tags:
- contract
- query
---

# Objective

    Add the fluent query interface to `/contract` (no implementation yet).


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `contract/query.go` defining:
      - builder methods: `Where/And/Or`, `Select`, `Resolve`, `OrderBy`, `Limit`
      - terminal methods: `List`, `Page`, `Stream`
    - `contract/page.go` with `PageResult` shape (minimal but usable)
    - `contract/iterator.go` defining `Iterator`
    - `contract/query_results.go` defining `QueryResults`

    ## Acceptance criteria

    - `go test ./contract` passes
    - No implementation details leak into contract
