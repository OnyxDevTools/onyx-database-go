package onyx

import "context"

// ListInto executes the query and decodes the results into dest.
// Dest should be a pointer to a slice or struct; Decode uses JSON tags on generated models.
func ListInto(ctx context.Context, q Query, dest any) error {
	results, err := q.List(ctx)
	if err != nil {
		return err
	}
	return results.Decode(dest)
}

// List executes the query and returns a fluent result that can be decoded.
func List(ctx context.Context, q Query) ListResult {
	results, err := q.List(ctx)
	return ListResult{results: results, err: err}
}

// ListResult is a fluent wrapper around query results for decoding.
type ListResult struct {
	results QueryResults
	err     error
}

// Decode populates dest (pointer to slice/struct) with the query results.
func (lr ListResult) Decode(dest any) error {
	if lr.err != nil {
		return lr.err
	}
	return lr.results.Decode(dest)
}
