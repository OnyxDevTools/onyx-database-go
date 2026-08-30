# v0.5.0

Intended release tag: `v0.5.0`

## Highlights

- Added a clean, typed `SearchWithOptions` API for lexical, semantic, and
  hybrid retrieval while preserving the existing `Search` behavior.
- Added strict validation and canonical wire serialization for search mode,
  term matching, minimum score, and candidate limits.

## Changes

- High-level search defaults to hybrid retrieval with any-term lexical
  matching and composes with ordinary structured filters.
- Search conditions are read-only and reject mutations, duplicate search
  roots, conflicting full-text predicates, and unsupported streams.
- Recursive condition inspection now enforces the same restrictions for
  custom compound conditions.
- Direct high-level searches can span table partitions under one global
  candidate budget; explicit partitions remain supported.
- Added documentation for embedding-provider compatibility, vector-space
  calibration, database-wide search scope, and the required resave/backfill of
  existing records before semantic retrieval.

## Compatibility

- This is a backward-compatible minor release.
- Existing `Search(query)` and `Search(query, minScore)` calls retain their
  legacy lexical `MATCHES` wire format. Callers opt into the new behavior with
  `SearchWithOptions`.
- The module remains tag-versioned; no runtime version constant is introduced.

## Release checks

- `go vet ./...`
- `go test ./... -coverprofile=coverage.out -covermode=atomic`
- `go build ./...`
