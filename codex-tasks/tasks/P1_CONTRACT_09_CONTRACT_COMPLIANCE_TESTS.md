---
id: P1_CONTRACT_09_CONTRACT_COMPLIANCE_TESTS
title: 'Add contract compliance tests: stdlib-only imports and API surface snapshot'
depends_on:
- P1_CONTRACT_02_ERRORS
- P1_CONTRACT_03_SORT
- P1_CONTRACT_04_CONDITIONS
- P1_CONTRACT_05_QUERY_INTERFACE
- P1_CONTRACT_06_CASCADE
- P1_CONTRACT_07_SCHEMA_TYPES
- P1_CONTRACT_08_CLIENT_INTERFACE
tags:
- contract
- tests
---

# Objective

    Add tests/guards that keep the contract lightweight and stable.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `contract/imports_test.go`:
      - assert that all `contract/*.go` files only import stdlib packages.
      - (implementation detail: parse Go files via `go/parser` and check import paths do not contain a dot.)
    - `contract/surface_test.go`:
      - snapshot high-level API surface in a simple text output and ensure it doesn't change unintentionally.
      - Keep the snapshot file in `contract/testdata/contract_surface.txt`.

    ## Acceptance criteria

    - `go test ./contract` passes
    - Failing tests clearly explain what changed
