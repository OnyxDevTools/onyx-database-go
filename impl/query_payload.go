package impl

import (
	"encoding/json"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type queryPayload struct {
	Type       string            `json:"type"`
	Table      string            `json:"table"`
	Fields     []string          `json:"fields,omitempty"`
	Conditions json.RawMessage   `json:"conditions,omitempty"`
	GroupBy    []string          `json:"groupBy,omitempty"`
	Resolvers  []string          `json:"resolvers,omitempty"`
	Sort       []json.RawMessage `json:"sort,omitempty"`
	Limit      *int              `json:"limit,omitempty"`
	Distinct   *bool             `json:"distinct,omitempty"`
	Partition  *string           `json:"partition,omitempty"`
}

func buildQueryPayload(q *query, includeLimit bool) queryPayload {
	payload := queryPayload{
		Type:       "SelectQuery",
		Table:      q.table,
		Fields:     nil,
		Conditions: nil,
		GroupBy:    nil,
		Resolvers:  nil,
		Sort:       nil,
		Limit:      nil,
		Distinct:   nil,
		Partition:  nil,
	}
	payload.Conditions = buildConditions(q.clauses)
	if q.partition != nil {
		payload.Partition = q.partition
	}

	if len(q.selectFields) > 0 {
		payload.Fields = append([]string{}, q.selectFields...)
	}
	if len(q.groupFields) > 0 {
		payload.GroupBy = append([]string{}, q.groupFields...)
	}
	if len(q.resolveFields) > 0 {
		payload.Resolvers = append([]string{}, q.resolveFields...)
	}
	if len(q.sorts) > 0 {
		for _, s := range q.sorts {
			raw, _ := json.Marshal(s)
			payload.Sort = append(payload.Sort, raw)
		}
	}
	if includeLimit && q.limit != nil {
		payload.Limit = q.limit
	}
	return payload
}

type updatePayload struct {
	Type       string            `json:"type"`
	Table      string            `json:"table"`
	Conditions json.RawMessage   `json:"conditions,omitempty"`
	Updates    map[string]any    `json:"updates"`
	Sort       []json.RawMessage `json:"sort,omitempty"`
	Limit      *int              `json:"limit,omitempty"`
	Partition  *string           `json:"partition,omitempty"`
}

func buildUpdatePayload(q *query) updatePayload {
	payload := updatePayload{
		Type:       "UpdateQuery",
		Table:      q.table,
		Conditions: buildConditions(q.clauses),
		Updates:    map[string]any{},
		Sort:       nil,
		Limit:      nil,
		Partition:  nil,
	}
	if q.partition != nil {
		payload.Partition = q.partition
	}
	for k, v := range q.updates {
		payload.Updates[k] = v
	}
	if len(q.sorts) > 0 {
		for _, s := range q.sorts {
			raw, _ := json.Marshal(s)
			payload.Sort = append(payload.Sort, raw)
		}
	}
	if q.limit != nil {
		payload.Limit = q.limit
	}
	return payload
}

func (p queryPayload) MarshalJSON() ([]byte, error) {
	type alias queryPayload
	return json.Marshal(alias(p))
}

func buildConditions(clauses []clause) json.RawMessage {
	if len(clauses) == 0 {
		return nil
	}

	buildSingle := func(c clause) map[string]any {
		raw, _ := json.Marshal(c.Condition)
		var m map[string]any
		_ = json.Unmarshal(raw, &m)
		return m
	}

	cur := buildSingle(clauses[0])
	for _, c := range clauses[1:] {
		cur = map[string]any{
			"conditionType": "CompoundCondition",
			"operator":      strings.ToUpper(c.Type),
			"conditions":    []any{cur, buildSingle(c)},
		}
	}

	out, _ := json.Marshal(cur)
	return out
}

var _ contract.Query = (*query)(nil)
