package contract

import (
	"encoding/json"
)

// Sort describes ordering for query results.
type Sort interface {
	json.Marshaler
}

type sortOrder struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

func (s sortOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Field     string `json:"field"`
		Direction string `json:"direction"`
	}{Field: s.Field, Direction: s.Direction})
}

// Asc sorts by the provided field in ascending order.
func Asc(field string) Sort {
	return sortOrder{Field: field, Direction: "asc"}
}

// Desc sorts by the provided field in descending order.
func Desc(field string) Sort {
	return sortOrder{Field: field, Direction: "desc"}
}
