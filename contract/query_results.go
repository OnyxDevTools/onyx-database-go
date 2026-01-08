package contract

import "encoding/json"

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

// Decode unmarshals the query results into the provided destination (pointer to slice/struct).
func (q QueryResults) Decode(dest any) error {
	raw, err := json.Marshal(q)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}
