---
id: P1_CONTRACT_06_CASCADE
title: Define cascade contract (CascadeSpec, CascadeBuilder, CascadeClient)
depends_on:
- P1_CONTRACT_01_PACKAGE_SKELETON
tags:
- contract
- cascade
---

# Objective

    Define cascade-related contract types, supporting:
    - string-based cascade spec: `contract.Cascade("userRoles:UserRole(userId,id)")`
    - builder-based creation for AI-friendliness


## Global constraints (apply to this task)

- Preserve the **contract-first** design: the `/contract` package is small, stable, and stdlib-only.
- **Do not** add non-stdlib dependencies to `/contract`.
- Any network/auth/caching/logging/resolvers belong in `/onyx` or `/internal`, never in `/contract`.
- Keep APIs **AI-friendly**: simple inputs, explicit outputs, deterministic behavior.
- Add unit tests with `go test ./...` passing.


    ## Required deliverables

    - `contract/cascade.go`
      - `type CascadeClient interface { Save(ctx, table string, entity any) error; Delete(ctx, table, id string) error }`
      - `type CascadeSpec interface { String() string }`
      - `type CascadeBuilder interface { Graph(name string) CascadeBuilder; GraphType(table string) CascadeBuilder; SourceField(field string) CascadeBuilder; TargetField(field string) CascadeBuilder; Build() CascadeSpec }`
      - `func Cascade(spec string) CascadeSpec`
      - `func NewCascadeBuilder() CascadeBuilder`
    - `contract/cascade_test.go`
      - builder emits same string as direct spec where applicable

    ## Acceptance criteria

    - `go test ./contract -run TestCascade` passes
    - Contract remains stdlib-only
