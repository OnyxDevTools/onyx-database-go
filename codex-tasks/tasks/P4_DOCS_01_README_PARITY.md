---
id: P4_DOCS_01_README_PARITY
title: Write README with Go equivalents of TS client examples (init/query/resolve/cascade/codegen)
depends_on:
- P3_SDK_11_CRUD_SAVE_DELETE
- P3_SDK_09_QUERY_LIST_PAGE
- P3_SDK_13_CASCADE_IMPL
- P2_CODEGEN_04_GENERATE_STRUCTS_AND_TABLES
- P2_SCHEMA_CLI_02_VALIDATE
tags:
- docs
---

# Objective

    Produce a high-quality `README.md`:
    - Explain contract-first layout
    - Show init via resolver and explicit config
    - Query builder examples mirroring TS
    - Resolve examples
    - Cascade examples
    - Schema CLI and codegen CLI usage


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Acceptance criteria

    - Examples compile (or are validated via tests/build tags)
    - Clear mapping from TS concepts to Go usage
