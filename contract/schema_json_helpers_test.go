package contract

import "testing"

func TestNormalizeResolversSortsAndConverts(t *testing.T) {
	schema := Schema{
		Tables: []Table{
			{
				Name: "User",
				Resolvers: []Resolver{
					{Name: "b"},
					{Name: "a"},
				},
			},
		},
	}
	normalized := NormalizeSchema(schema)
	if len(normalized.Tables[0].Resolvers) != 2 || normalized.Tables[0].Resolvers[0].Name != "a" {
		t.Fatalf("expected resolvers normalized and sorted, got %+v", normalized.Tables[0].Resolvers)
	}

	// Exercise schemaFromTablesArray branches (indexes, triggers, meta, resolvers)
	raw := []byte(`{
  "tables": [{
    "name": "Doc",
    "fields": [{"name":"id","type":"String","nullable":false,"primaryKey":true,"unique":true}],
    "resolvers": [{"name":"r","resolver":"db.from(\"X\")","meta":{"a":1}}],
    "indexes": [{"name":"idx"}],
    "triggers": [{"name":"tr"}],
    "meta": {"k":"v"}
  }]
}`)
	parsed, err := ParseSchemaJSON(raw)
	if err != nil {
		t.Fatalf("parse schema json: %v", err)
	}
	if len(parsed.Tables) != 1 || parsed.Tables[0].Indexes[0].Name != "idx" || parsed.Tables[0].Triggers[0] != "tr" {
		t.Fatalf("expected indexes/triggers/meta parsed: %+v", parsed.Tables[0])
	}
}
