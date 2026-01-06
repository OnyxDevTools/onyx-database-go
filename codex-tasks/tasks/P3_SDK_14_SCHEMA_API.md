---
id: P3_SDK_14_SCHEMA_API
title: Implement Client.Schema(ctx) fetching schema from API
depends_on:
- P3_SDK_05_HTTP_CLIENT_CORE
- P3_SDK_06_AUTH_SIGNING
- P3_SDK_07_ERROR_MAPPING
- P1_CONTRACT_07_SCHEMA_TYPES
- P1_CONTRACT_08_CLIENT_INTERFACE
tags:
- sdk
- schema
- api
---

# Objective

    Implement `Client.Schema(ctx)` which fetches schema from the Onyx API and returns `contract.Schema`.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required work

    - Identify TS endpoint and response JSON shape.
    - Parse into contract schema types using `contract.ParseSchemaJSON` (or a dedicated API parser).

    ## Required deliverables

    - `onyx/schema_api.go`
    - tests with fixture response bodies

    ## Acceptance criteria

    - Returned schema is normalized/deterministic
