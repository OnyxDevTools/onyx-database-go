# onyx-database-go Codex Task Pack

This zip contains a dependency-aware task list for building `onyx-database-go` (Go SDK + CLIs) using a **contract-first** approach.

## Contents

- `tasks.yaml` — machine-readable manifest with dependencies
- `tasks/*.md` — one task per file, with YAML frontmatter
- Tasks are named with phase prefixes:
  - `P0_...` setup
  - `P1_...` contract (must be frozen before implementation)
  - `P2_...` local-only CLIs (validate/diff/codegen from file)
  - `P3_...` SDK implementation
  - `P4_...` integrations, docs, CI, release, parity

## Concurrency

Any tasks that do not depend on each other can be executed concurrently.
In general:
- After `P1_CONTRACT_01_PACKAGE_SKELETON`, most `P1_CONTRACT_*` tasks can run in parallel.
- `P2_SCHEMA_CLI_*` and `P2_CODEGEN_*` can run in parallel once the contract is frozen.
- `P3_SDK_02_RESOLVER_*` and `P3_SDK_05_HTTP_CLIENT_CORE` can run in parallel.

## Notes

This pack does not assume any particular task-runner implementation.
If your runner supports:
- reading `tasks.yaml`
- respecting `depends_on`
- running tasks in parallel

…then you can start immediately.

If your runner only supports markdown, you can still run tasks manually in the suggested order.
