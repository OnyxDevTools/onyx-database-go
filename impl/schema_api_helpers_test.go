package impl

import (
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestSchemaHelperConversions(t *testing.T) {
	entities := []any{
		map[string]any{
			"name": "User",
			"identifier": map[string]any{
				"name": "id",
			},
			"attributes": []any{
				map[string]any{"name": "id", "type": "String", "isNullable": false},
				map[string]any{"name": "age", "type": "Int", "isNullable": true},
			},
			"resolvers": []any{"roles"},
		},
	}
	schema := schemaFromEntities(entities)
	if len(schema.Tables) != 1 || schema.Tables[0].Fields[0].Primary != true {
		t.Fatalf("expected primary key detected: %+v", schema.Tables[0].Fields)
	}
	if len(schema.Tables[0].Resolvers) != 1 || schema.Tables[0].Resolvers[0].Name != "roles" {
		t.Fatalf("resolver not parsed: %+v", schema.Tables[0].Resolvers)
	}

	tablesArray := []any{
		map[string]any{
			"name": "Doc",
			"fields": []any{
				map[string]any{"name": "id", "type": "String", "nullable": false, "primaryKey": true},
			},
		},
	}
	schema2 := schemaFromTablesArray(tablesArray)
	if len(schema2.Tables) != 1 || !schema2.Tables[0].Fields[0].Primary {
		t.Fatalf("expected primary key from tables array")
	}

	// coverage for helper wrappers
	_ = stringValue("x")
	_ = boolValue(true)
	_ = mapValue(map[string]any{"a": 1})
}

func TestToEntitiesIncludesResolversMeta(t *testing.T) {
	schema := contract.Schema{
		Tables: []contract.Table{
			{
				Name: "T",
				Fields: []contract.Field{
					{Name: "id", Type: "String", Primary: true},
					{Name: "n", Type: "String", Nullable: true},
				},
				Resolvers: []contract.Resolver{
					{Name: "r", Resolver: "db.from(\"X\")", Meta: map[string]any{"a": 1}},
				},
				Indexes: []contract.Index{{Name: "idx"}},
				Triggers: []string{
					"trg",
				},
				Meta: map[string]any{"m": "v"},
			},
		},
	}
	entities := toEntities(schema)
	if len(entities) != 1 || len(entities[0].Resolvers) != 1 {
		t.Fatalf("expected resolver exported: %+v", entities)
	}
	if len(entities[0].Indexes) != 1 || len(entities[0].Triggers) != 1 || entities[0].Meta["m"] != "v" {
		t.Fatalf("expected indexes/triggers/meta exported: %+v", entities[0])
	}
}
