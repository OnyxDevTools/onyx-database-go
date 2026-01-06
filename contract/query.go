package contract

import "context"

// Query defines the fluent query builder interface used by the SDK.
type Query interface {
	Where(condition Condition) Query
	And(condition Condition) Query
	Or(condition Condition) Query
	Select(fields ...string) Query
	Resolve(paths ...string) Query
	OrderBy(sorts ...Sort) Query
	Limit(limit int) Query

	List(ctx context.Context) (QueryResults, error)
	Page(ctx context.Context, cursor string) (PageResult, error)
	Stream(ctx context.Context) (Iterator, error)

	MarshalJSON() ([]byte, error)
}
