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
func (s stubQuery) Resolve(...string) Query                    { return s }
func (s stubQuery) OrderBy(...Sort) Query                      { return s }
func (s stubQuery) Limit(int) Query                            { return s }
func (s stubQuery) List(context.Context) (QueryResults, error) { return nil, nil }
func (s stubQuery) Page(context.Context, string) (PageResult, error) {
	return PageResult{}, nil
}
func (s stubQuery) Stream(context.Context) (Iterator, error) { return nil, nil }
func (s stubQuery) MarshalJSON() ([]byte, error)             { return []byte(`{"table":"User"}`), nil }

func TestConditionJSON(t *testing.T) {
	sampleQuery := stubQuery{}

	cases := []struct {
		name string
		cond Condition
		want string
	}{
		{name: "eq", cond: Eq("name", "alice"), want: `{"field":"name","op":"eq","value":"alice"}`},
		{name: "neq", cond: Neq("age", 30), want: `{"field":"age","op":"neq","value":30}`},
		{name: "in", cond: In("role", []any{"admin", "member"}), want: `{"field":"role","op":"in","values":["admin","member"]}`},
		{name: "not_in", cond: NotIn("role", []any{"guest"}), want: `{"field":"role","op":"not_in","values":["guest"]}`},
		{name: "between", cond: Between("score", 1, 10), want: `{"field":"score","from":1,"op":"between","to":10}`},
		{name: "gt", cond: Gt("score", 9), want: `{"field":"score","op":"gt","value":9}`},
		{name: "gte", cond: Gte("score", 9), want: `{"field":"score","op":"gte","value":9}`},
		{name: "lt", cond: Lt("score", 9), want: `{"field":"score","op":"lt","value":9}`},
		{name: "lte", cond: Lte("score", 9), want: `{"field":"score","op":"lte","value":9}`},
		{name: "like", cond: Like("email", "%@example.com"), want: `{"field":"email","op":"like","pattern":"%@example.com"}`},
		{name: "contains", cond: Contains("tags", "blue"), want: `{"field":"tags","op":"contains","value":"blue"}`},
		{name: "starts_with", cond: StartsWith("name", "Al"), want: `{"field":"name","op":"starts_with","value":"Al"}`},
		{name: "is_null", cond: IsNull("deletedAt"), want: `{"field":"deletedAt","op":"is_null"}`},
		{name: "not_null", cond: NotNull("createdAt"), want: `{"field":"createdAt","op":"not_null"}`},
		{name: "within", cond: Within("userId", sampleQuery), want: `{"field":"userId","op":"within","query":{"table":"User"}}`},
		{name: "not_within", cond: NotWithin("userId", sampleQuery), want: `{"field":"userId","op":"not_within","query":{"table":"User"}}`},
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
