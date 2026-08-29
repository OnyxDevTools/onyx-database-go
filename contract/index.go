package contract

// IndexType identifies how an Onyx schema index is managed.
type IndexType string

const (
	// IndexTypeDefault is an ordinary secondary index.
	IndexTypeDefault IndexType = "DEFAULT"
	// IndexTypeVector is a native vector-managed index on a searchable table.
	IndexTypeVector IndexType = "VECTOR"
)

// Index describes an index definition.
type Index struct {
	Name string    `json:"name"`
	Type IndexType `json:"type,omitempty"`
	// MinimumScore is retained for schema round trips. Score thresholds are
	// query-time controls and current servers ignore this schema field.
	MinimumScore *float64 `json:"minimumScore,omitempty"`
}
