---
id: P3_SDK_02_RESOLVER_ENV
title: 'Implement config resolver: explicit + env vars'
depends_on:
- P3_SDK_01_PUBLIC_INIT
tags:
- sdk
- resolver
---

# Objective

    Implement a resolver chain in `onyx/resolver` supporting:
    - explicit config values (highest precedence)
    - environment variables:
      - `ONYX_DATABASE_ID`
      - `ONYX_DATABASE_BASE_URL`
      - `ONYX_DATABASE_API_KEY`
      - `ONYX_DATABASE_API_SECRET`


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `onyx/resolver/resolve.go`
      - `Resolve(ctx, partial Config) (ResolvedConfig, Meta, error)`
      - Meta includes `Source` per field (useful for debug logs)
    - tests for precedence and missing required fields

    ## Acceptance criteria

    - explicit overrides env
    - env-only works when complete
