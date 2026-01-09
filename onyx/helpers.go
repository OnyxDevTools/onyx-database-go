package onyx

import (
	"context"

	"github.com/OnyxDevTools/onyx-database-go/core"
)

// ListInto executes the query and decodes the results into dest.
// Dest should be a pointer to a slice or struct; Decode uses JSON tags on generated models.
func ListInto(ctx context.Context, q Query, dest any) error {
	return core.ListInto(ctx, q, dest)
}

// List executes the query and returns a fluent result that can be decoded.
func List(ctx context.Context, q Query) ListResult {
	return core.List(ctx, q)
}
