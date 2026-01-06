---
id: P3_SDK_07_ERROR_MAPPING
title: Map non-2xx HTTP responses into contract.Error consistently
depends_on:
- P3_SDK_05_HTTP_CLIENT_CORE
- P1_CONTRACT_02_ERRORS
tags:
- sdk
- errors
---

# Objective

    Ensure all SDK errors are returned as `*contract.Error` (or wrap it).


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `internal/httpclient/errors.go`:
      - parse error response bodies (if JSON) to fill `Code`, `Message`, `Meta`
      - always include HTTP status in Meta
    - tests:
      - JSON error body
      - non-JSON error body
      - context cancellation

    ## Acceptance criteria

    - Callers can type-assert to `*contract.Error` for API failures
