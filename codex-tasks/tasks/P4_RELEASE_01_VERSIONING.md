---
id: P4_RELEASE_01_VERSIONING
title: Define release/versioning process for Go module + CLIs
depends_on:
- P4_DOCS_01_README_PARITY
- P4_CI_01_COVERAGE_GATE
tags:
- release
---

# Objective

    Define a release process:
    - semantic versioning for module tags
    - changelog or release notes
    - how to publish binary releases for CLIs (optional)


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Deliverables

    - `RELEASING.md` with steps
    - (optional) `.github/workflows/release.yml` draft

    ## Acceptance criteria

    - Steps are deterministic and reproducible
