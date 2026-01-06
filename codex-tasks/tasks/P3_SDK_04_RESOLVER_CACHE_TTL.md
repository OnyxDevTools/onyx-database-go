---
id: P3_SDK_04_RESOLVER_CACHE_TTL
title: Add resolver caching (default 5m) + ClearConfigCache parity
depends_on:
- P3_SDK_03_RESOLVER_FILES
tags:
- sdk
- resolver
- cache
---

# Objective

    Implement resolved-config caching with a default TTL of 5 minutes, configurable via `Config`.

    Expose `onyx.ClearConfigCache()` that clears the cache.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `onyx/resolver/cache.go` with a threadsafe cache
    - tests verifying:
      - cache hit within TTL
      - cache miss after TTL
      - clear forces reload

    ## Acceptance criteria

    - deterministic tests (use injected clock or time control)
