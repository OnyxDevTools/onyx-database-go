# v0.6.0

Intended release tag: `v0.6.0`

## Highlights

- MessagePack is now the default transport for entity saves, deletes, batches,
  queries, and query streams; JSON remains available as an explicit compatibility
  mode.
- Added per-table `SearchSupportLexical`, `SearchSupportSemantic`, and
  `SearchSupportBoth` schema capabilities with lossless Cloud schema round trips.
- Added `Table.SearchSupport` to the stable public contract and re-exported the
  type and constants from the `onyx` facade.

## Changes

- An empty `Config.WireFormat` now selects `WireFormatMessagePack`. Set it to
  `WireFormatJSON` to retain the previous entity transport. Documents, schemas,
  secrets, configuration discovery, and AI routes continue to use JSON.
- Entity-shaped and table-shaped schema parsing now retain `searchSupport`, and
  schema publish payloads send an explicitly configured value back to Cloud.
- Schema diff output now reports table management-type and effective search
  capability changes. Omitted capabilities remain equivalent to `BOTH`.
- Documentation and contract snapshots include lexical-only, semantic-only,
  and combined searchable-table configuration.

## Compatibility

- This pre-1.0 minor release changes the default entity wire format from JSON to
  MessagePack. Applications that require JSON must opt in with `WireFormatJSON`.
- Public client method and configuration field signatures are unchanged.
- Existing schema values with no `searchSupport` continue to mean `BOTH`, and
  the field has no effect on ordinary tables.
- Existing query and high-level search APIs are unchanged.
- The module remains tag-versioned; no runtime version constant is introduced.

## Release checks

- `go vet ./...`
- `go test ./... -coverprofile=coverage.out -covermode=atomic`
- `go build ./...`
