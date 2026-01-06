---
id: P3_SDK_09_QUERY_LIST_PAGE
title: 'Wire query builder to API: List and Page terminal methods'
depends_on:
- P3_SDK_08_QUERY_BUILDER_IMPL
- P3_SDK_05_HTTP_CLIENT_CORE
- P3_SDK_06_AUTH_SIGNING
- P3_SDK_07_ERROR_MAPPING
tags:
- sdk
- query
- api
---

# Objective

    Implement `Query.List(ctx)` and `Query.Page(ctx)` by calling the Onyx API endpoints.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required work

    - Inspect TS client to find:
      - endpoint paths
      - request body shape
      - response shape for list/page and pagination tokens
    - Implement response decoding into:
      - `contract.QueryResults` (wrap underlying list of maps)
      - `contract.PageResult`

    ## Required deliverables

    - `onyx/query_exec.go`
    - `onyx/query_results_impl.go`
    - tests using `httptest.Server` with fixture responses

    ## Acceptance criteria

    - Behavior matches TS (especially paging fields)
