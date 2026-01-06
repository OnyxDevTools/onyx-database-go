---
id: P4_PARITY_01_TS_AUDIT
title: 'Perform parity audit vs TypeScript SDK: endpoints, JSON shapes, query/cascade
  behaviors'
depends_on:
- P3_SDK_12_BATCH_SAVE
- P3_SDK_10_QUERY_STREAM
- P4_ENDPOINTS_01_DOCUMENTS_API
- P4_ENDPOINTS_02_SECRETS_API
- P4_SCHEMA_CLI_01_GET_PUBLISH
- P4_CODEGEN_01_SOURCE_API
tags:
- parity
- audit
---

# Objective

    Create a parity audit document comparing Go SDK vs TS SDK.


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `PARITY.md` with checklist:
      - init + resolver chain order
      - cache TTL + clear
      - request logging toggles
      - condition/sort JSON encoding
      - query endpoints + pagination
      - streaming behavior
      - cascade behavior
      - documents/secrets/schema APIs
      - codegen outputs and ergonomics

    ## Acceptance criteria

    - All non-parity items are explicitly documented with rationale and follow-up tasks
