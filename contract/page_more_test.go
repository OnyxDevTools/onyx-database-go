package contract

import "testing"

func TestParseSchemaJSONDirectTables(t *testing.T) {
	schema, err := ParseSchemaJSON([]byte(`{"tables":[{"name":"X","fields":[{"name":"id","type":"String"}]}]}`))
	if err != nil || len(schema.Tables) != 1 {
		t.Fatalf("expected direct tables parse, got err=%v schema=%+v", err, schema)
	}
}

func TestQueryResultsArrayShape(t *testing.T) {
	var q QueryResults
	if err := q.UnmarshalJSON([]byte(`[{"id":1}]`)); err != nil {
		t.Fatalf("expected array shape decode: %v", err)
	}
	if len(q) != 1 {
		t.Fatalf("expected one element")
	}
}
