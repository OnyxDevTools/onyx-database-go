package contract

import "encoding/json"

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

func (c condition) MarshalJSON() ([]byte, error) {
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
