---
id: P4_SCHEMA_CLI_01_GET_PUBLISH
title: Implement onyx-schema-go get/publish using onyx SDK resolver + HTTP stack
depends_on:
- P3_SDK_14_SCHEMA_API
- P3_SDK_04_RESOLVER_CACHE_TTL
- P2_SCHEMA_CLI_01_SKELETON
tags:
- cli
- schema
- api
---

# Objective

    Add API-backed commands to `onyx-schema-go`:
    - `get` (fetch schema and write to stdout or `--out`)
    - `publish` (read local schema and publish)


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Requirements

    - Reuse `onyx.InitWithDatabaseID` + resolver chain.
    - Flags:
      - `--database-id` optional if resolver can infer
      - `--out` path for `get`
      - `--schema` path for `publish` (default `./onyx.schema.json`)
    - Provide `--json` output mode for errors where feasible.

    ## Acceptance criteria

    - Commands compile and have unit tests (mock http server)
