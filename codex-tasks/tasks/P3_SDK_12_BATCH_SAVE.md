---
id: P3_SDK_12_BATCH_SAVE
title: Implement Client.BatchSave with chunking and safe retries
depends_on:
- P3_SDK_11_CRUD_SAVE_DELETE
tags:
- sdk
- api
- batch
---

# Objective

    Implement batch save with chunking similar to TS `batchSave`.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Requirements

    - `batchSize` default behavior:
      - if <=0, use a sensible default (e.g., 500) and document it
    - Must:
      - chunk entities
      - call underlying API per chunk
      - stop on first non-retryable error
    - Retry:
      - only on transient status codes (e.g., 429/502/503/504) with bounded backoff
      - deterministic tests (inject sleeper)

    ## Required deliverables

    - `onyx/batch.go`
    - tests covering chunking and retry boundaries

    ## Acceptance criteria

    - Correct number of API calls for N entities
