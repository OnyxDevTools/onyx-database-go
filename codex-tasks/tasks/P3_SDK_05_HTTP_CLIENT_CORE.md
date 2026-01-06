---
id: P3_SDK_05_HTTP_CLIENT_CORE
title: 'Implement internal HTTP client: base URL, keepalive transport, logging toggles'
depends_on:
- P3_SDK_01_PUBLIC_INIT
tags:
- sdk
- http
---

# Objective

    Create internal HTTP client with:
    - base URL handling
    - tuned `http.Transport` keep-alive pooling
    - request/response logging toggles
      - explicit config flags
      - `ONYX_DEBUG=true` enables both logs


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `internal/httpclient/client.go`
      - `DoJSON(ctx, method, path string, reqBody any, respBody any) error`
      - redaction for secrets in logs
    - tests using `httptest.Server` for:
      - JSON encoding/decoding
      - logging on/off (can assert via injected logger)

    ## Acceptance criteria

    - No logs by default
    - `ONYX_DEBUG=true` turns on logs
