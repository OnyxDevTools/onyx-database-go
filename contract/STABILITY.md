# Contract stability policy

The `onyx` package defines the API surface shared across the Onyx Database Go SDK. It is stdlib-only and intentionally free of runtime behavior so the API remains deterministic and easy to audit. Runtime logic lives in `impl` (and `internal` helpers), which depends on `onyx`. Renaming or collapsing these packages would break the contract-first boundary and is considered a breaking change. We follow semantic versioning with additive-only changes permitted in v1.x.

## Compatible changes
- Adding new exported types, interfaces, constants, or helper constructors.
- Adding optional fields to structs when existing behavior is preserved.
- Expanding documentation or comments without changing semantics.

## Breaking changes (require a major version bump)
- Removing or renaming exported packages, types, functions, methods, or fields.
- Changing method signatures, parameter types, return types, or exported struct field types.
- Altering default behaviors, invariants, or JSON shapes in a way that changes existing consumers' expectations.
- Introducing dependencies outside the Go standard library into the `onyx` package.
- Reordering or mutating stable identifiers in a way that would break serialization contracts.

## Process
Breaking changes must be documented here alongside the rationale and carried with a major version release. Additive changes should include tests to lock in behavior and keep the package deterministic.
