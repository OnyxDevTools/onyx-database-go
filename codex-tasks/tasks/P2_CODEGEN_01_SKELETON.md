---
id: P2_CODEGEN_01_SKELETON
title: Create onyx-gen-go CLI skeleton (schema file -> Go file)
depends_on:
- P1_CONTRACT_10_FREEZE_V1
tags:
- cli
- codegen
---

# Objective

    Implement the CLI skeleton for `onyx-gen-go` which generates Go types from `onyx.schema.json`.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Flags (v1)

    - `--schema` (default `./onyx.schema.json`)
    - `--out` (required)
    - `--package` (required)
    - `--tables` (optional comma-separated allowlist)
    - `--timestamps` (`time` or `string`, default `time`)

    ## Required deliverables

    - `cmd/onyx-gen-go/main.go`
    - `cmd/onyx-gen-go/generator/*` package with a clear entrypoint
    - `go run ./cmd/onyx-gen-go --help` works

    ## Acceptance criteria

    - CLI shows help and validates required flags
