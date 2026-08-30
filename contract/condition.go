package contract

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Condition represents a filter operator used in a query.
type Condition interface {
	json.Marshaler
}

type queryProvider interface {
	MarshalJSON() ([]byte, error)
}

type condition struct {
	op     string
	field  string
	value  any
	values []any
	from   any
	to     any
	query  queryProvider
}

// FullTextQuery is the legacy text-only MATCHES payload. New semantic and
// hybrid callers should use VectorSearchQuery.
type FullTextQuery struct {
	QueryText string   `json:"queryText"`
	MinScore  *float64 `json:"minScore"`
}

const fullTextField = "__full_text__"

func (c condition) MarshalJSON() ([]byte, error) {
	if c.CandidateOperator() != "" && strings.TrimSpace(c.field) == "" {
		return nil, fmt.Errorf("candidate attribute must not be blank")
	}
	crit := map[string]any{
		"field":    c.field,
		"operator": operatorFor(c.op),
	}

	switch c.op {
	case "in", "not_in":
		if c.values != nil {
			crit["value"] = c.values
		}
	case "between":
		crit["value"] = map[string]any{"from": c.from, "to": c.to}
	case "is_null", "not_null":
		// no value
	case "within", "not_within":
		if c.query != nil {
			raw, err := c.query.MarshalJSON()
			if err != nil {
				return nil, err
			}
			crit["value"] = json.RawMessage(raw)
		}
	default:
		crit["value"] = c.value
	}

	return json.Marshal(map[string]any{
		"conditionType": "SingleCondition",
		"criteria":      crit,
	})
}

// CandidateOperator reports the wire operator when this condition must be the
// sole root query criterion. It lets SDK query builders enforce the invariant
// without expanding the stable Condition interface.
func (c condition) CandidateOperator() string {
	switch c.op {
	case "candidates", "search_candidates", "hnsw_candidates":
		return operatorFor(c.op)
	default:
		return ""
	}
}

// ReadOnlyOperator reports the wire operator when a condition is valid only
// for read queries. Unlike CandidateOperator, it does not imply that the
// condition must be the sole root criterion.
func (c condition) ReadOnlyOperator() string {
	if candidate := c.CandidateOperator(); candidate != "" {
		return candidate
	}
	if c.op == "search" {
		return operatorFor(c.op)
	}
	return ""
}

func operatorFor(op string) string {
	switch op {
	case "eq":
		return "EQUAL"
	case "neq":
		return "NOT_EQUAL"
	case "in":
		return "IN"
	case "not_in":
		return "NOT_IN"
	case "between":
		return "BETWEEN"
	case "gt":
		return "GREATER_THAN"
	case "gte":
		return "GREATER_THAN_EQUAL"
	case "lt":
		return "LESS_THAN"
	case "lte":
		return "LESS_THAN_EQUAL"
	case "like":
		return "LIKE"
	case "contains":
		return "CONTAINS"
	case "starts_with":
		return "STARTS_WITH"
	case "is_null":
		return "IS_NULL"
	case "not_null":
		return "NOT_NULL"
	case "within":
		return "IN"
	case "not_within":
		return "NOT_IN"
	case "matches":
		return "MATCHES"
	case "search":
		return "SEARCH"
	case "candidates":
		return "CANDIDATES"
	case "search_candidates":
		return "SEARCH_CANDIDATES"
	case "hnsw_candidates":
		return "HNSW_CANDIDATES"
	default:
		return op
	}
}

// Eq creates an equality condition for the given field.
func Eq(field string, value any) Condition { return condition{op: "eq", field: field, value: value} }

// Neq creates an inequality condition for the given field.
func Neq(field string, value any) Condition { return condition{op: "neq", field: field, value: value} }

// In creates a membership condition for the given field.
func In(field string, values []any) Condition {
	return condition{op: "in", field: field, values: values}
}

// NotIn creates a negated membership condition for the given field.
func NotIn(field string, values []any) Condition {
	return condition{op: "not_in", field: field, values: values}
}

// Between creates a range condition for the given field.
func Between(field string, from, to any) Condition {
	return condition{op: "between", field: field, from: from, to: to}
}

// Gt creates a greater-than condition.
func Gt(field string, value any) Condition { return condition{op: "gt", field: field, value: value} }

// Gte creates a greater-than-or-equal condition.
func Gte(field string, value any) Condition { return condition{op: "gte", field: field, value: value} }

// Lt creates a less-than condition.
func Lt(field string, value any) Condition { return condition{op: "lt", field: field, value: value} }

// Lte creates a less-than-or-equal condition.
func Lte(field string, value any) Condition { return condition{op: "lte", field: field, value: value} }

// Like matches a value using a pattern.
func Like(field string, pattern any) Condition {
	return condition{op: "like", field: field, value: pattern}
}

// Contains matches containers that include a value.
func Contains(field string, value any) Condition {
	return condition{op: "contains", field: field, value: value}
}

// StartsWith matches string prefixes.
func StartsWith(field string, value any) Condition {
	return condition{op: "starts_with", field: field, value: value}
}

// Search performs a full-text search across all indexed fields.
// When minScore is omitted, it is serialized as null.
func Search(queryText string, minScore ...float64) Condition {
	var score *float64
	if len(minScore) > 0 {
		v := minScore[0]
		score = &v
	}
	return condition{
		op:    "matches",
		field: fullTextField,
		value: FullTextQuery{QueryText: queryText, MinScore: score},
	}
}

// SearchWithOptions creates a high-level lexical, semantic, or hybrid search
// condition. It may be composed with ordinary filters, but queries containing
// it are read-only.
func SearchWithOptions(queryText string, options SearchOptions) Condition {
	return condition{
		op:    "search",
		field: fullTextField,
		value: newTextSearchQuery(queryText, options),
	}
}

// VectorSearch creates a native lexical, semantic, or hybrid MATCHES condition.
// Unlike bounded candidate operators, this condition may be composed with
// ordinary filters.
func VectorSearch(searchQuery VectorSearchQuery) Condition {
	return condition{
		op:    "matches",
		field: fullTextField,
		value: searchQuery,
	}
}

// ApproximateSearch creates a bounded lexical candidate condition. It must be
// the query's sole root criterion and is valid only for read operations.
func ApproximateSearch(searchQuery VectorSearchQuery) Condition {
	return condition{
		op:    "search_candidates",
		field: fullTextField,
		value: boundedLexicalSearchQuery{query: searchQuery},
	}
}

// HNSWCandidates creates a bounded native-HNSW nearest-neighbor condition. It
// must be the query's sole root criterion and is valid only for read operations.
func HNSWCandidates(searchQuery HNSWSearchQuery) Condition {
	return condition{
		op:    "hnsw_candidates",
		field: fullTextField,
		value: searchQuery,
	}
}

// ApproximateCandidates creates a bounded ordinary-index candidate condition.
// A scalar produces EQUAL-style routing; a slice or array produces IN-style
// routing. It must be the query's sole root criterion and is read-only.
func ApproximateCandidates(field string, valueOrValues any, maxCandidates ...int) Condition {
	query, err := NewApproximateIndexCandidateQuery(valueOrValues, maxCandidates...)
	var value any = query
	if err != nil {
		value = invalidSearchValue{err: err}
	}
	return condition{op: "candidates", field: field, value: value}
}

type boundedLexicalSearchQuery struct {
	query VectorSearchQuery
}

func (q boundedLexicalSearchQuery) MarshalJSON() ([]byte, error) {
	normalized, err := q.query.normalized()
	if err != nil {
		return nil, err
	}
	if normalized.Text == nil || normalized.Semantic != nil {
		return nil, fmt.Errorf("SEARCH_CANDIDATES supports text-only VectorSearchQuery values")
	}
	return json.Marshal(normalized)
}

type invalidSearchValue struct {
	err error
}

func (v invalidSearchValue) MarshalJSON() ([]byte, error) {
	return nil, v.err
}

// IsNull checks for null values.
func IsNull(field string) Condition { return condition{op: "is_null", field: field} }

// NotNull checks for non-null values.
func NotNull(field string) Condition { return condition{op: "not_null", field: field} }

// Within matches values found in a nested query.
func Within(field string, query Query) Condition {
	return condition{op: "within", field: field, query: query}
}

// NotWithin excludes values found in a nested query.
func NotWithin(field string, query Query) Condition {
	return condition{op: "not_within", field: field, query: query}
}
