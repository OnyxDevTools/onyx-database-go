//go:generate env ONYX_GEN_TIMESTAMP=1970-01-01T00:00:00Z go run ./cmd/onyx-go gen --source file --schema ./examples/api/onyx.schema.json --out ./examples/gen/onyx --package onyx

package tools

// This file anchors go generate for deterministic example client output.
