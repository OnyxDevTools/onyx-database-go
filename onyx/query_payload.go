package onyx

import (
	"encoding/json"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type queryPayload struct {
	Table   string                       `json:"table"`
	Where   map[string][]json.RawMessage `json:"where,omitempty"`
	Select  []string                     `json:"select,omitempty"`
	Resolve []string                     `json:"resolve,omitempty"`
	OrderBy []json.RawMessage            `json:"orderBy,omitempty"`
	Limit   int                          `json:"limit,omitempty"`
}

func buildQueryPayload(q *query) queryPayload {
	payload := queryPayload{Table: q.table}
	if len(q.clauses) > 0 {
		payload.Where = map[string][]json.RawMessage{}
		for _, clause := range q.clauses {
			raw, _ := json.Marshal(clause.Condition)
			payload.Where[clause.Type] = append(payload.Where[clause.Type], raw)
		}
	}

	if len(q.selectFields) > 0 {
		payload.Select = append([]string{}, q.selectFields...)
	}
	if len(q.resolveFields) > 0 {
		payload.Resolve = append([]string{}, q.resolveFields...)
	}
	if len(q.sorts) > 0 {
		for _, s := range q.sorts {
			raw, _ := json.Marshal(s)
			payload.OrderBy = append(payload.OrderBy, raw)
		}
	}
	if q.limit > 0 {
		payload.Limit = q.limit
	}
	return payload
}

func (p queryPayload) MarshalJSON() ([]byte, error) {
	type alias queryPayload
	return json.Marshal(alias(p))
}

var _ contract.Query = (*query)(nil)
