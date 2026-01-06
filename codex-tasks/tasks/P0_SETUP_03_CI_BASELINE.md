---
id: P0_SETUP_03_CI_BASELINE
title: Add GitHub Actions CI for go test and basic checks
depends_on:
- P0_SETUP_01_REPO_SCAFFOLD
tags:
- setup
- ci
---

# Objective

    Add baseline CI in `.github/workflows/ci.yml` that runs on PRs and pushes:
    - `go test ./...`
    - `go vet ./...`


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Notes

    - Keep it minimal; coverage gates come later.
    - Use the repo's Go version from `go.mod`.

    ## Acceptance criteria

    - Workflow file exists and is syntactically valid.
    - `go test ./...` passes in CI.
