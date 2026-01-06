---
id: P0_SETUP_01_REPO_SCAFFOLD
title: Repo scaffold for onyx-database-go (dirs, go.mod, baseline files)
depends_on: []
tags:
- setup
- foundation
---

# Objective

    Create the initial repository scaffold for **onyx-database-go** with a clean Go layout that supports:
    - a stable `/contract` package
    - an implementation `/onyx` package
    - CLI tools under `/cmd`
    - internal-only helpers under `/internal`


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    1. `go.mod` with module path `github.com/OnyxDevTools/onyx-database-go`
    2. Directory structure (create empty packages with `doc.go` as needed):
       - `contract/`
       - `onyx/`
       - `internal/`
       - `cmd/onyx-schema-go/`
       - `cmd/onyx-gen-go/`
       - `examples/`
    3. Baseline repo files:
       - `LICENSE` (Apache-2.0 to match typical OpenAI/Onyx style)
       - `.gitignore` (Go + OS)
       - `README.md` placeholder describing contract-first approach and CLIs (high-level only)

    ## Acceptance criteria

    - `go test ./...` passes (even if no tests exist yet)
    - `go list ./...` works with no errors
    - `/contract` compiles with stdlib only (no external deps)
