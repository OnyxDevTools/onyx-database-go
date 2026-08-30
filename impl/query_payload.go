package impl

import (
	"encoding/json"
	"fmt"
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

func buildQueryPayload(q *query, includeLimit bool) (queryPayload, error) {
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
	if q.buildErr != nil {
		return queryPayload{}, q.buildErr
	}
	conditions, err := buildConditions(q.clauses)
	if err != nil {
		return queryPayload{}, err
	}
	if err := validateConditionPlan(conditions); err != nil {
		return queryPayload{}, err
	}
	payload.Conditions = conditions
	if q.partition != nil {
		payload.Partition = q.partition
	}

	if len(q.selectFields) > 0 {
		payload.Fields = append([]string{}, q.selectFields...)
	}
	if len(q.groupFields) > 0 {
		payload.GroupBy = append([]string{}, q.groupFields...)
	}
	if q.distinct {
		distinct := true
		payload.Distinct = &distinct
	}
	if len(q.resolveFields) > 0 {
		payload.Resolvers = append([]string{}, q.resolveFields...)
	}
	if len(q.sorts) > 0 {
		for _, s := range q.sorts {
			raw, err := json.Marshal(s)
			if err != nil {
				return queryPayload{}, fmt.Errorf("marshal query sort: %w", err)
			}
			payload.Sort = append(payload.Sort, raw)
		}
	}
	if includeLimit && q.limit != nil {
		payload.Limit = q.limit
	}
	return payload, nil
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

func buildUpdatePayload(q *query) (updatePayload, error) {
	if q.buildErr != nil {
		return updatePayload{}, q.buildErr
	}
	conditions, err := buildConditions(q.clauses)
	if err != nil {
		return updatePayload{}, err
	}
	if err := validateConditionPlan(conditions); err != nil {
		return updatePayload{}, err
	}
	payload := updatePayload{
		Type:       "UpdateQuery",
		Table:      q.table,
		Conditions: conditions,
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
			raw, err := json.Marshal(s)
			if err != nil {
				return updatePayload{}, fmt.Errorf("marshal update sort: %w", err)
			}
			payload.Sort = append(payload.Sort, raw)
		}
	}
	if q.limit != nil {
		payload.Limit = q.limit
	}
	return payload, nil
}

func (p queryPayload) MarshalJSON() ([]byte, error) {
	type alias queryPayload
	return json.Marshal(alias(p))
}

func buildConditions(clauses []clause) (json.RawMessage, error) {
	if len(clauses) == 0 {
		return nil, nil
	}

	buildSingle := func(c clause) (map[string]any, error) {
		raw, err := json.Marshal(c.Condition)
		if err != nil {
			return nil, fmt.Errorf("marshal query condition: %w", err)
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("decode query condition: %w", err)
		}
		return m, nil
	}

	cur, err := buildSingle(clauses[0])
	if err != nil {
		return nil, err
	}
	for _, c := range clauses[1:] {
		next, err := buildSingle(c)
		if err != nil {
			return nil, err
		}
		cur = map[string]any{
			"conditionType": "CompoundCondition",
			"operator":      strings.ToUpper(c.Type),
			"conditions":    []any{cur, next},
		}
	}

	out, err := json.Marshal(cur)
	if err != nil {
		return nil, fmt.Errorf("marshal query conditions: %w", err)
	}
	return out, nil
}

var _ contract.Query = (*query)(nil)
