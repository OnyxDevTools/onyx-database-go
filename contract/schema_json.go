package contract

import (
	"encoding/json"
	"sort"
)

// ParseSchemaJSON parses a schema document from JSON bytes.
func ParseSchemaJSON(data []byte) (Schema, error) {
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return Schema{}, err
	}
	return s, nil
}

// NormalizeSchema returns a copy of the schema with deterministic ordering.
func NormalizeSchema(s Schema) Schema {
	normalized := Schema{Tables: make([]Table, len(s.Tables))}
	copy(normalized.Tables, s.Tables)

	sort.Slice(normalized.Tables, func(i, j int) bool {
		return normalized.Tables[i].Name < normalized.Tables[j].Name
	})

	for i := range normalized.Tables {
		fields := make([]Field, len(normalized.Tables[i].Fields))
		copy(fields, normalized.Tables[i].Fields)
		sort.Slice(fields, func(a, b int) bool {
			return fields[a].Name < fields[b].Name
		})
		normalized.Tables[i].Fields = fields
	}

	return normalized
}
