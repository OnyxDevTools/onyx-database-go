---
id: P4_ENDPOINTS_01_DOCUMENTS_API
title: Implement Documents API surface in onyx SDK (parity with TS docs)
depends_on:
- P3_SDK_05_HTTP_CLIENT_CORE
- P3_SDK_06_AUTH_SIGNING
- P3_SDK_07_ERROR_MAPPING
- P3_SDK_01_PUBLIC_INIT
tags:
- sdk
- api
- documents
---

# Objective

    Implement the Documents API surface that exists in the TS client / Onyx docs.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required work

    - Identify endpoints and payloads from TS client and/or docs.
    - Implement Go methods on `onyx` client (implementation package), keeping contract stable.
    - Add tests for each endpoint.

    ## Acceptance criteria

    - Feature parity with TS for Documents operations
