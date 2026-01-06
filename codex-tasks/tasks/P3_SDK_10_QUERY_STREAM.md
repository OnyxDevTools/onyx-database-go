---
id: P3_SDK_10_QUERY_STREAM
title: Implement Query.Stream(ctx) iterator using streaming results API
depends_on:
- P3_SDK_09_QUERY_LIST_PAGE
tags:
- sdk
- query
- streaming
---

# Objective

    Implement streaming query results in Go.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required work

    - Determine the server streaming protocol from TS client / docs:
      - chunked JSON? NDJSON? SSE?
    - Implement `Iterator` that:
      - respects context cancellation
      - closes response body
      - returns stable errors

    ## Required deliverables

    - `onyx/query_stream.go`
    - tests that simulate streaming responses via `httptest.Server`

    ## Acceptance criteria

    - No goroutine leaks in tests (use `-race` locally if possible)
