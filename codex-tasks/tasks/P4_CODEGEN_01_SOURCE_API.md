---
id: P4_CODEGEN_01_SOURCE_API
title: Add onyx-gen-go --source api to fetch schema via SDK before generating
depends_on:
- P3_SDK_14_SCHEMA_API
- P2_CODEGEN_01_SKELETON
- P2_CODEGEN_02_LOAD_SCHEMA
tags:
- cli
- codegen
- api
---

# Objective

    Extend `onyx-gen-go` so it can load schema from the API:

    - `--source file|api` (default file)
    - When source=api:
      - use resolver/init
      - call `Client.Schema(ctx)`
      - generate from the returned schema


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Acceptance criteria

    - Codegen behavior is identical between file vs api when schemas match
