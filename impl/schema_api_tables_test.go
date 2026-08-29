package impl

import (
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestSchemaFromTablesArrayCoversAllBranches(t *testing.T) {
	raw := []any{
		map[string]any{
			"name": "Doc",
			"type": "SEARCHABLE",
			"fields": []any{
				map[string]any{"name": "id", "type": "String", "nullable": false, "primaryKey": true, "unique": true},
				map[string]any{"name": "note", "type": "String", "nullable": true},
			},
			"resolvers": []any{
				"plain",
				map[string]any{"name": "withMeta", "resolver": "db.from(\"Doc\")", "meta": map[string]any{"k": "v"}},
			},
			"indexes": []any{
				map[string]any{"name": "content", "type": "VECTOR"},
			},
		},
	}

	schema := schemaFromTablesArray(raw)
	if len(schema.Tables) != 1 {
		t.Fatalf("expected one table, got %d", len(schema.Tables))
	}
	tbl := schema.Tables[0]
	if !tbl.Fields[0].Primary || !tbl.Fields[0].Unique {
		t.Fatalf("expected flags set on first field: %+v", tbl.Fields[0])
	}
	if len(tbl.Resolvers) != 2 || tbl.Resolvers[1].Meta["k"] != "v" {
		t.Fatalf("resolver meta missing: %+v", tbl.Resolvers)
	}
	if tbl.Type != contract.TableTypeSearchable || len(tbl.Indexes) != 1 || tbl.Indexes[0].Type != contract.IndexTypeVector {
		t.Fatalf("native vector schema metadata missing: %+v", tbl)
	}
	if len(tbl.Triggers) != 0 || tbl.Meta != nil {
		t.Fatalf("unexpected extras parsed: %+v %+v", tbl.Triggers, tbl.Meta)
	}
}
