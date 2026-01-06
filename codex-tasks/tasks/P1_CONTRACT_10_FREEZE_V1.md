---
id: P1_CONTRACT_10_FREEZE_V1
title: Freeze contract v1 checklist and tag readiness notes
depends_on:
- P1_CONTRACT_09_CONTRACT_COMPLIANCE_TESTS
tags:
- contract
- release
---

# Objective

    Create a short checklist that must be green before starting SDK implementation work.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `contract/V1_FREEZE_CHECKLIST.md` including:
      - all contract tests passing
      - no external deps
      - stable JSON shapes for conditions/sorts
      - semver policy written
      - review notes: what assumptions were made about API JSON shapes and where to verify them (TS client / docs)

    ## Acceptance criteria

    - Checklist exists and is actionable
