---
id: P3_SDK_11_CRUD_SAVE_DELETE
title: Implement Client.Save and Client.Delete API calls
depends_on:
- P3_SDK_05_HTTP_CLIENT_CORE
- P3_SDK_06_AUTH_SIGNING
- P3_SDK_07_ERROR_MAPPING
- P3_SDK_01_PUBLIC_INIT
- P1_CONTRACT_08_CLIENT_INTERFACE
tags:
- sdk
- api
- crud
---

# Objective

    Implement `contract.Client` methods:
    - `Save(ctx, table, entity)`
    - `Delete(ctx, table, id)`


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required work

    - Find exact endpoint paths/payloads in TS client.
    - Ensure entity encoding preserves arbitrary fields (map/struct).

    ## Required deliverables

    - `onyx/crud.go`
    - tests with `httptest.Server`

    ## Acceptance criteria

    - Matches TS response expectations (if any)
