---
id: P3_SDK_13_CASCADE_IMPL
title: Implement CascadeClient.Save/Delete using cascade spec
depends_on:
- P3_SDK_11_CRUD_SAVE_DELETE
- P1_CONTRACT_06_CASCADE
tags:
- sdk
- cascade
- api
---

# Objective

    Implement cascade operations matching TS client.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required work

    - Determine cascade endpoints and payload shapes from TS client.
    - Implement `Client.Cascade(spec).Save(...)` and `.Delete(...)`.

    ## Required deliverables

    - `onyx/cascade.go`
    - tests using `httptest.Server`

    ## Acceptance criteria

    - Spec string is passed exactly as produced by `contract.CascadeSpec.String()`
