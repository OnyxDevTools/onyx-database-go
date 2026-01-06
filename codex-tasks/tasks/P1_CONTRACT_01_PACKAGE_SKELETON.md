---
id: P1_CONTRACT_01_PACKAGE_SKELETON
title: Create contract package skeleton and stability policy doc
depends_on:
- P0_SETUP_01_REPO_SCAFFOLD
- P0_SETUP_02_AGENT_GUIDANCE
tags:
- contract
- foundation
---

# Objective

    Create the `/contract` package skeleton and a clear **stability policy**.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `contract/doc.go` with package docs describing:
      - purpose of contract
      - stdlib-only constraint
      - semver rules (additive-only in v1)
    - `contract/STABILITY.md` describing what is considered breaking.

    ## Acceptance criteria

    - `go test ./contract` passes
    - `/contract` imports stdlib only
