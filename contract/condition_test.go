package contract

import (
	"context"
	"encoding/json"
	"testing"
)

type stubQuery struct{}

func (s stubQuery) Where(Condition) Query                      { return s }
func (s stubQuery) And(Condition) Query                        { return s }
func (s stubQuery) Or(Condition) Query                         { return s }
func (s stubQuery) Select(...string) Query                     { return s }
func (s stubQuery) GroupBy(...string) Query                    { return s }
func (s stubQuery) Resolve(...string) Query                    { return s }
func (s stubQuery) OrderBy(...Sort) Query                      { return s }
func (s stubQuery) Limit(int) Query                            { return s }
func (s stubQuery) List(context.Context) (QueryResults, error) { return nil, nil }
func (s stubQuery) Delete(context.Context) (int, error)        { return 0, nil }
func (s stubQuery) Page(context.Context, string) (PageResult, error) {
	return PageResult{}, nil
}
func (s stubQuery) Stream(context.Context) (Iterator, error) { return nil, nil }
func (s stubQuery) SetUpdates(map[string]any) Query          { return s }
func (s stubQuery) Update(context.Context) (int, error)      { return 0, nil }
func (s stubQuery) InPartition(string) Query                 { return s }
func (s stubQuery) MarshalJSON() ([]byte, error)             { return []byte(`{"table":"User"}`), nil }

func TestConditionJSON(t *testing.T) {
	sampleQuery := stubQuery{}

	cases := []struct {
		name string
		cond Condition
		want string
	}{
		{name: "eq", cond: Eq("name", "alice"), want: `{"conditionType":"SingleCondition","criteria":{"field":"name","operator":"EQUAL","value":"alice"}}`},
		{name: "neq", cond: Neq("age", 30), want: `{"conditionType":"SingleCondition","criteria":{"field":"age","operator":"NOT_EQUAL","value":30}}`},
		{name: "in", cond: In("role", []any{"admin", "member"}), want: `{"conditionType":"SingleCondition","criteria":{"field":"role","operator":"IN","value":["admin","member"]}}`},
		{name: "not_in", cond: NotIn("role", []any{"guest"}), want: `{"conditionType":"SingleCondition","criteria":{"field":"role","operator":"NOT_IN","value":["guest"]}}`},
		{name: "between", cond: Between("score", 1, 10), want: `{"conditionType":"SingleCondition","criteria":{"field":"score","operator":"BETWEEN","value":{"from":1,"to":10}}}`},
		{name: "gt", cond: Gt("score", 9), want: `{"conditionType":"SingleCondition","criteria":{"field":"score","operator":"GREATER_THAN","value":9}}`},
		{name: "gte", cond: Gte("score", 9), want: `{"conditionType":"SingleCondition","criteria":{"field":"score","operator":"GREATER_THAN_EQUAL","value":9}}`},
		{name: "lt", cond: Lt("score", 9), want: `{"conditionType":"SingleCondition","criteria":{"field":"score","operator":"LESS_THAN","value":9}}`},
		{name: "lte", cond: Lte("score", 9), want: `{"conditionType":"SingleCondition","criteria":{"field":"score","operator":"LESS_THAN_EQUAL","value":9}}`},
		{name: "like", cond: Like("email", "%@example.com"), want: `{"conditionType":"SingleCondition","criteria":{"field":"email","operator":"LIKE","value":"%@example.com"}}`},
		{name: "contains", cond: Contains("tags", "blue"), want: `{"conditionType":"SingleCondition","criteria":{"field":"tags","operator":"CONTAINS","value":"blue"}}`},
		{name: "starts_with", cond: StartsWith("name", "Al"), want: `{"conditionType":"SingleCondition","criteria":{"field":"name","operator":"STARTS_WITH","value":"Al"}}`},
		{name: "is_null", cond: IsNull("deletedAt"), want: `{"conditionType":"SingleCondition","criteria":{"field":"deletedAt","operator":"IS_NULL"}}`},
		{name: "not_null", cond: NotNull("createdAt"), want: `{"conditionType":"SingleCondition","criteria":{"field":"createdAt","operator":"NOT_NULL"}}`},
		{name: "within", cond: Within("userId", sampleQuery), want: `{"conditionType":"SingleCondition","criteria":{"field":"userId","operator":"IN","value":{"table":"User"}}}`},
		{name: "not_within", cond: NotWithin("userId", sampleQuery), want: `{"conditionType":"SingleCondition","criteria":{"field":"userId","operator":"NOT_IN","value":{"table":"User"}}}`},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.cond)
			if err != nil {
				t.Fatalf("marshal condition: %v", err)
			}

			if string(data) != tt.want {
				t.Fatalf("unexpected json. got=%s want=%s", string(data), tt.want)
			}
		})
	}
}
