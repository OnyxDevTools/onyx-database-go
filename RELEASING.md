# Releasing onyx-database-go

This repository follows semantic versioning for the module and publishes reproducible CLI binaries.

## Prepare a release
1. Ensure `main` is green: `go vet ./...`, `go test ./... -coverprofile=coverage.out -covermode=atomic`.
2. Update `README.md` and examples if any public APIs changed.
3. Add a changelog entry to the GitHub release notes (see template below).
4. Tag the release with `vX.Y.Z` and push the tag.

## Tagging
```
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

## GitHub Release workflow (optional)
If desired, add a `release.yml` workflow that builds binaries for `cmd/onyx-go` and `cmd/onyx-schema-go` on tags and uploads them as artifacts. Binaries must be deterministic and produced via `go build ./cmd/...`.

## Release notes template
```
## Highlights
- ...

## Changes
- ...

## Checks
- go vet ./...
- go test ./... -coverprofile=coverage.out -covermode=atomic
```
