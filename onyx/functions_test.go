package onyx

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type stubMarshalQuery struct{}

func (s stubMarshalQuery) Where(condition contract.Condition) contract.Query       { return s }
func (s stubMarshalQuery) And(condition contract.Condition) contract.Query         { return s }
func (s stubMarshalQuery) Or(condition contract.Condition) contract.Query          { return s }
func (s stubMarshalQuery) Select(fields ...string) contract.Query                  { return s }
func (s stubMarshalQuery) GroupBy(fields ...string) contract.Query                 { return s }
func (s stubMarshalQuery) Resolve(paths ...string) contract.Query                  { return s }
func (s stubMarshalQuery) OrderBy(sorts ...contract.Sort) contract.Query           { return s }
func (s stubMarshalQuery) Limit(limit int) contract.Query                          { return s }
func (s stubMarshalQuery) SetUpdates(updates map[string]any) contract.Query        { return s }
func (s stubMarshalQuery) MarshalJSON() ([]byte, error)                            { return []byte(`{"query":"ok"}`), nil }
func (s stubMarshalQuery) List(ctx context.Context) (contract.QueryResults, error) { return nil, nil }
func (s stubMarshalQuery) Page(ctx context.Context, cursor string) (contract.PageResult, error) {
	return contract.PageResult{}, nil
}
func (s stubMarshalQuery) Stream(ctx context.Context) (contract.Iterator, error) { return nil, nil }
func (s stubMarshalQuery) Update(ctx context.Context) (int, error)               { return 0, nil }
func (s stubMarshalQuery) Delete(ctx context.Context) (int, error)               { return 0, nil }
func (s stubMarshalQuery) InPartition(string) contract.Query                     { return s }

func TestReExportedHelpers(t *testing.T) {
	assertJSONEqual := func(t *testing.T, got, want any) {
		t.Helper()
		gotJSON, err := json.Marshal(got)
		if err != nil {
			t.Fatalf("marshal got: %v", err)
		}
		wantJSON, err := json.Marshal(want)
		if err != nil {
			t.Fatalf("marshal want: %v", err)
		}
		if string(gotJSON) != string(wantJSON) {
			t.Fatalf("expected %s, got %s", string(wantJSON), string(gotJSON))
		}
	}

	assertJSONEqual(t, Asc("name"), contract.Asc("name"))
	assertJSONEqual(t, Desc("name"), contract.Desc("name"))
	assertJSONEqual(t, Eq("field", 1), contract.Eq("field", 1))
	assertJSONEqual(t, Neq("field", 1), contract.Neq("field", 1))
	assertJSONEqual(t, In("field", []any{1, 2}), contract.In("field", []any{1, 2}))
	assertJSONEqual(t, NotIn("field", []any{1, 2}), contract.NotIn("field", []any{1, 2}))
	assertJSONEqual(t, Between("field", 1, 2), contract.Between("field", 1, 2))
	assertJSONEqual(t, Gt("field", 1), contract.Gt("field", 1))
	assertJSONEqual(t, Gte("field", 1), contract.Gte("field", 1))
	assertJSONEqual(t, Lt("field", 1), contract.Lt("field", 1))
	assertJSONEqual(t, Lte("field", 1), contract.Lte("field", 1))
	assertJSONEqual(t, Like("field", "x"), contract.Like("field", "x"))
	assertJSONEqual(t, Contains("field", "x"), contract.Contains("field", "x"))
	assertJSONEqual(t, StartsWith("field", "x"), contract.StartsWith("field", "x"))
	assertJSONEqual(t, IsNull("field"), contract.IsNull("field"))
	assertJSONEqual(t, NotNull("field"), contract.NotNull("field"))

	query := stubMarshalQuery{}
	assertJSONEqual(t, Within("field", query), contract.Within("field", query))
	assertJSONEqual(t, NotWithin("field", query), contract.NotWithin("field", query))

	cascade := Cascade("graph:User")
	if cascade.String() != "graph:User" {
		t.Fatalf("expected cascade spec, got %s", cascade.String())
	}

	builder := NewCascadeBuilder().Graph("g").GraphType("User").SourceField("id").TargetField("user_id")
	if builder.Build().String() != "g:User(id,user_id)" {
		t.Fatalf("unexpected cascade builder output: %s", builder.Build().String())
	}

	sdkErr := NewError("code", "message", map[string]any{"a": 1})
	if sdkErr.Code != "code" || sdkErr.Message != "message" {
		t.Fatalf("unexpected error fields: %+v", sdkErr)
	}

	schema := Schema{Tables: []Table{{Name: "b", Fields: []Field{{Name: "id", Type: "string"}}}, {Name: "a", Fields: []Field{{Name: "id", Type: "string"}}}}}
	normalized := NormalizeSchema(schema)
	if len(normalized.Tables) != 2 || normalized.Tables[0].Name != "a" {
		t.Fatalf("expected normalized schema, got %+v", normalized.Tables)
	}

	parsed, err := ParseSchemaJSON([]byte(`{"tables":[{"name":"Users","fields":[]}]}`))
	if err != nil || len(parsed.Tables) != 1 {
		t.Fatalf("expected parsed schema, got %+v (%v)", parsed, err)
	}

}

func TestInitWrappers(t *testing.T) {
	cfg := Config{
		DatabaseID:      "db",
		DatabaseBaseURL: "https://example.com",
		APIKey:          "key",
		APISecret:       "secret",
	}

	client, err := Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if client == nil {
		t.Fatalf("expected client")
	}

	t.Setenv("ONYX_DATABASE_BASE_URL", "https://example.com")
	t.Setenv("ONYX_DATABASE_API_KEY", "key")
	t.Setenv("ONYX_DATABASE_API_SECRET", "secret")

	client, err = InitWithDatabaseID(context.Background(), "db")
	if err != nil {
		t.Fatalf("InitWithDatabaseID error: %v", err)
	}
	if client == nil {
		t.Fatalf("expected client")
	}

	ClearConfigCache()
}
