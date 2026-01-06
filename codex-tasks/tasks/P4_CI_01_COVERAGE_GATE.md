---
id: P4_CI_01_COVERAGE_GATE
title: Add CI coverage gate and enforce minimum threshold
depends_on:
- P0_SETUP_03_CI_BASELINE
- P3_SDK_11_CRUD_SAVE_DELETE
tags:
- ci
- tests
---

# Objective

    Add a coverage gate in CI. Choose a threshold that is strict but achievable.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Requirements

    - Generate coverage profile in CI
    - Fail PRs that reduce coverage below threshold
    - Document how to run coverage locally

    ## Acceptance criteria

    - CI fails when coverage drops below threshold
