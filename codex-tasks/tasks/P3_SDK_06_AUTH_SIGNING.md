---
id: P3_SDK_06_AUTH_SIGNING
title: Implement API auth/signing exactly like TS client
depends_on:
- P3_SDK_05_HTTP_CLIENT_CORE
- P3_SDK_02_RESOLVER_ENV
tags:
- sdk
- http
- auth
---

# Objective

    Implement request authentication/signing identical to the TypeScript client.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required work

    - Inspect the TypeScript client in `@onyx.dev/onyx-database` to determine:
      - which headers are required (api key, secret, signatures, timestamps, etc.)
      - canonical request format and hash algorithms if any
      - how request IDs are generated
    - Implement in `internal/httpclient/auth.go`.

    ## Required deliverables

    - `internal/httpclient/auth.go`
    - unit tests:
      - given fixed inputs (method/path/body/timestamp), signature matches expected
      - ensure deterministic behavior via injected clock / nonce generator

    ## Acceptance criteria

    - Matches TS behavior byte-for-byte for the same request inputs
