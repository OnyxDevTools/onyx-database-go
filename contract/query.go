package contract

import "context"

// Query defines the fluent query builder interface used by the SDK.
type Query interface {
	Where(condition Condition) Query
	And(condition Condition) Query
	Or(condition Condition) Query
	Select(fields ...string) Query
	GroupBy(fields ...string) Query
	Resolve(paths ...string) Query
	OrderBy(sorts ...Sort) Query
	Limit(limit int) Query
	SetUpdates(updates map[string]any) Query
	Update(ctx context.Context) (int, error)

	List(ctx context.Context) (QueryResults, error)
	Page(ctx context.Context, cursor string) (PageResult, error)
	Stream(ctx context.Context) (Iterator, error)

	MarshalJSON() ([]byte, error)
}
