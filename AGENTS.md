# Codex Guidance for onyx-database-go

- Contract-first: `/contract` is stdlib-only and stable. No network/auth/cache/logging/resolver imports. Do not break the contract without a deliberate major bump and updating `contract/STABILITY.md`.
- Boundaries: runtime behavior lives in `/onyx` or `/internal`; CLIs and generated code must never import `/internal`.
- Generated artifacts: keep codegen deterministic (stable ordering, no timestamps), import `contract` only (plus `time` when timestamps are emitted), and ensure outputs compile standalone. Commit generated files only when explicitly requested.
- Determinism: keep condition/sort JSON shapes stable, normalize schemas before diff/codegen, and ensure CLI output ordering is reproducible.
- Testing: every change in `/onyx`, `/cmd`, or `/internal` requires unit tests. Keep contract compliance tests (stdlib-only imports, surface snapshot) green. Run `go test ./...` and `go vet ./...`; run `golangci-lint run` when available.
- Resolver/HTTP parity: resolver precedence is explicit config > env vars > config files with TTL caching and `ClearConfigCache`; HTTP client defaults to quiet logging unless `ONYX_DEBUG=true`; auth/signing must match the TS client byte-for-byte.
- CLIs: use stdlib flag parsing, exit codes 0 (success) / 1 (failure) / 2 (usage). Schema/codegen commands must produce deterministic output and honor the contract-only import rule.
