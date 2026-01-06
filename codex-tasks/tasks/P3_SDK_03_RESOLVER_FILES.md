---
id: P3_SDK_03_RESOLVER_FILES
title: 'Implement config resolver: ONYX_CONFIG_PATH + well-known JSON paths'
depends_on:
- P3_SDK_02_RESOLVER_ENV
tags:
- sdk
- resolver
---

# Objective

    Add file-based config resolution matching TS behavior:

    - If `ONYX_CONFIG_PATH` is set, load that JSON first (after explicit, before other files).
    - Otherwise search (in this order):
      - `./onyx-database-<databaseId>.json`
      - `./onyx-database.json`
      - `~/.onyx/onyx-database-<databaseId>.json`
      - `~/.onyx/onyx-database.json`
      - `~/onyx-database.json`


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - JSON config struct matches TS fields: `databaseId`, `databaseBaseUrl`, `apiKey`, `apiSecret`
    - tests using temp dirs and fake HOME

    ## Acceptance criteria

    - correct search order
    - databaseId-specific file wins over generic when both exist
