package contract

import (
	"encoding/json"
)

// Sort describes ordering for query results.
type Sort interface {
	json.Marshaler
}

type sortOrder struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

func (s sortOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Field string `json:"field"`
		Order string `json:"order"`
	}{Field: s.Field, Order: s.Order})
}

// Asc sorts by the provided field in ascending order.
func Asc(field string) Sort {
	return sortOrder{Field: field, Order: "ASC"}
}

// Desc sorts by the provided field in descending order.
func Desc(field string) Sort {
	return sortOrder{Field: field, Order: "DESC"}
}
