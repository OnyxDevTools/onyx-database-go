---
id: P0_SETUP_02_AGENT_GUIDANCE
title: Add AGENTS.md guidance for Codex work in this repo
depends_on:
- P0_SETUP_01_REPO_SCAFFOLD
tags:
- setup
- codex
---

# Objective

    Add a repository-root `AGENTS.md` with guidance that optimizes Codex performance for this repo.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required content in AGENTS.md

    - Emphasize **contract-first** and that `/contract` is stdlib-only and stable.
    - Emphasize no breaking changes in `/contract` without a deliberate major bump.
    - Require tests for any behavior in `/onyx`, `/cmd`, `/internal`.
    - Require deterministic codegen outputs (stable ordering, no timestamps).
    - Require that generated code imports `contract` only.
    - Require running:
      - `go test ./...`
      - `go vet ./...`
      - (optional) `golangci-lint run` if present

    ## Acceptance criteria

    - `AGENTS.md` exists at repo root and is concise and actionable.
