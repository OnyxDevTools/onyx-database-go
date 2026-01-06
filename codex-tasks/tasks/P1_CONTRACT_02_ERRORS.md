---
id: P1_CONTRACT_02_ERRORS
title: Define contract.Error and error conversion expectations
depends_on:
- P1_CONTRACT_01_PACKAGE_SKELETON
tags:
- contract
---

# Objective

    Define a minimal, stable error model in `/contract` to wrap all SDK/CLI errors.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `contract/errors.go`:
      - `type Error struct { Code string; Message string; Meta map[string]any }`
      - implement `Error() string`
      - helper `func NewError(code, message string, meta map[string]any) *Error`
    - `contract/errors_test.go`:
      - stable string formatting
      - nil meta safe

    ## Acceptance criteria

    - `go test ./contract -run TestError` passes
    - No external deps
