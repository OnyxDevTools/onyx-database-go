package contract

import (
	"encoding/json"
	"fmt"
)

// QueryResults represents a collection of query rows.
type QueryResults []map[string]any

// UnmarshalJSON supports both array responses and objects of the form
// { "records": [...] } returned by the API.
func (q *QueryResults) UnmarshalJSON(data []byte) error {
	var items []map[string]any
	if err := json.Unmarshal(data, &items); err == nil {
		*q = items
		return nil
	}

	var wrapper struct {
		Records []map[string]any `json:"records"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	*q = wrapper.Records
	return nil
}

// UnmarshalMessagePackValue consumes the validated JSON-shaped value tree
// produced by the entity MessagePack decoder. It deliberately avoids
// re-marshaling through encoding/json so dynamic signed integers remain int64.
func (q *QueryResults) UnmarshalMessagePackValue(value any) error {
	if wrapper, ok := value.(map[string]any); ok {
		value = wrapper["records"]
	}

	if value == nil {
		*q = nil
		return nil
	}

	items, ok := value.([]any)
	if !ok {
		return fmt.Errorf("msgpack: cannot decode query results from %T", value)
	}
	results := make(QueryResults, len(items))
	for i, item := range items {
		if item == nil {
			continue
		}
		row, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("msgpack: query result at index %d is %T; expected an object", i, item)
		}
		results[i] = row
	}
	*q = results
	return nil
}

// Decode unmarshals the query results into the provided destination (pointer to slice/struct).
func (q QueryResults) Decode(dest any) error {
	raw, err := json.Marshal(q)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}
