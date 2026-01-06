---
id: P2_SCHEMA_CLI_01_SKELETON
title: Create onyx-schema-go CLI skeleton with subcommand dispatch
depends_on:
- P1_CONTRACT_10_FREEZE_V1
tags:
- cli
- schema
---

# Objective

    Implement the CLI skeleton for `onyx-schema-go` with subcommands:
    - `validate`
    - `diff`
    - (stubs) `get`, `publish` (wired later when SDK HTTP stack exists)


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Requirements

    - Use stdlib-only CLI parsing (`os.Args` + `flag` packages).
    - Provide `--help` output.
    - Provide consistent exit codes: 0 success, 2 usage, 1 failure.

    ## Required deliverables

    - `cmd/onyx-schema-go/main.go`
    - `cmd/onyx-schema-go/commands/*.go` (or similar minimal structure)
    - Ensure `go test ./...` passes

    ## Acceptance criteria

    - `go run ./cmd/onyx-schema-go --help` works
    - `go run ./cmd/onyx-schema-go validate --help` works
