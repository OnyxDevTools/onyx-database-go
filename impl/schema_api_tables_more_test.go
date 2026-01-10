package impl

import (
	"testing"
)

func TestSchemaFromEntitiesCoverage(t *testing.T) {
	raw := []any{
		map[string]any{
			"name": "Entity",
			"attributes": []any{
				map[string]any{"name": "id", "type": "String"},
			},
			"identifier": map[string]any{"name": "id"},
			"resolvers": []any{
				map[string]any{"name": "r", "resolver": "code", "meta": map[string]any{"a": 1}},
			},
			"meta": map[string]any{"owner": "team"},
		},
	}
	schema := schemaFromEntities(raw)
	if len(schema.Tables) != 1 || schema.Tables[0].Fields[0].Primary != true {
		t.Fatalf("expected primary id")
	}
}
