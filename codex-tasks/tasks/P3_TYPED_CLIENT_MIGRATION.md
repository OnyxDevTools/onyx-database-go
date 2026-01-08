Title: Typed client API migration plan (single import, db.User() style)

Goal: let callers use only `onyx` and write `db.User().Save(...)` without breaking existing two-import flows.

Phases
- Phase 1: Break circular deps. Extract a minimal core (query/condition/sort helpers) that depends only on contract; re-export from `onyx`. Point the generator to this core instead of `onyx` so generated code no longer imports `onyx`.
- Phase 2: Typed façade on the core client. Extend `contract.Client`/`impl.client` with a `Typed()` accessor that wraps the generated client and exposes table services: `User()`, `Role()`, etc. Keep `onyxclient.NewClient` for compatibility.
- Phase 3: One-import ergonomics. Add a helper on `onyx.Client` to return the typed façade so callers do `db.Typed().User().Save(...)`. Old API remains usable.
- Phase 4: Migrate examples. Update examples to the new style; leave notes on legacy usage during transition.
- Phase 5: Deprecate legacy entrypoints. Document and eventually retire the two-import pattern (`onyxclient.NewClient`) after a deprecation window.

Notes
- Keep contract surface stable; only add to it.
- No behavioral regressions during migration; old paths must keep working until deprecation.
