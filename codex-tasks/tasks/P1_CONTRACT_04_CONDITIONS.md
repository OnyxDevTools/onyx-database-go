---
id: P1_CONTRACT_04_CONDITIONS
title: Define contract.Condition and condition helper constructors
depends_on:
- P1_CONTRACT_01_PACKAGE_SKELETON
tags:
- contract
- query
---

# Objective

    Implement a **pure-data** condition algebra in `/contract`:
    - `Condition` interface with `MarshalJSON()`
    - helper constructors for common operators:
      `Eq, Neq, In, NotIn, Between, Gt, Gte, Lt, Lte, Like, Contains, StartsWith, IsNull, NotNull, Within, NotWithin`


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Implementation requirements

    - Each condition should marshal to the exact JSON shape required by the Onyx API.
    - For `Within/NotWithin`, represent the subquery in a way that can later be serialized by the implementation.
      - Accept `contract.Query` as input (do not import implementation).
      - Use a small internal struct that stores `json.RawMessage` or a `QueryMarshaler` interface.

    ## Required deliverables

    - `contract/condition.go`
    - `contract/condition_test.go` with golden JSON tests for every operator
    - Ensure deterministic ordering inside JSON objects if applicable

    ## Acceptance criteria

    - `go test ./contract -run TestCondition` passes
    - No external deps
