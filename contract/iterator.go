package contract

// Iterator provides streaming access to query results.
type Iterator interface {
	Next() bool
	Value() map[string]any
	Err() error
	Close() error
}
