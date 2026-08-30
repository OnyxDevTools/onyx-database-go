# v0.6.0

Intended release tag: `v0.6.0`

## Highlights

- Added per-table `SearchSupportLexical`, `SearchSupportSemantic`, and
  `SearchSupportBoth` schema capabilities with lossless Cloud schema round trips.
- Added `Table.SearchSupport` to the stable public contract and re-exported the
  type and constants from the `onyx` facade.

## Changes

- Entity-shaped and table-shaped schema parsing now retain `searchSupport`, and
  schema publish payloads send an explicitly configured value back to Cloud.
- Schema diff output now reports table management-type and effective search
  capability changes. Omitted capabilities remain equivalent to `BOTH`.
- Documentation and contract snapshots include lexical-only, semantic-only,
  and combined searchable-table configuration.

## Compatibility

- This is a backward-compatible minor release.
- Existing schema values with no `searchSupport` continue to mean `BOTH`, and
  the field has no effect on ordinary tables.
- Existing query and high-level search APIs are unchanged.
- The module remains tag-versioned; no runtime version constant is introduced.

## Release checks

- `go vet ./...`
- `go test ./... -coverprofile=coverage.out -covermode=atomic`
- `go build ./...`
