# Go SDK vs TypeScript SDK parity audit

The Go SDK mirrors the TypeScript client surface. This checklist records observed parity; any gaps include rationale and follow-up.

- [x] Init + resolver precedence (explicit > env vars > config files) with `ClearConfigCache` support
- [x] Cache TTL + explicit `ClearConfigCache` hook to reset resolver state between calls
- [x] Request logging toggles (`LogRequests`, `LogResponses`, and `ONYX_DEBUG` override)
- [x] Condition and sort JSON encoding mirrors TS shapes (see contract tests)
- [x] Query endpoints: list, page, stream implemented with pagination cursors
- [x] Streaming behavior implemented via incremental scanner with error propagation
- [x] Cascade behavior matches contract cascade spec helpers
- [x] Documents API: list/get/save/delete parity helpers
- [x] Secrets API: list/get/set/delete parity helpers
- [x] Schema API: fetch plus publish endpoints used by CLI and codegen
- [x] Codegen outputs deterministic structs/helpers with API and file schema parity

No outstanding non-parity items are known at this time; new endpoints should extend this checklist.
