package contract

import (
	"encoding/json"
	"testing"
)

func TestPageResultUnmarshalVariants(t *testing.T) {
	var p PageResult
	if err := json.Unmarshal([]byte(`{"items":[{"id":1}],"nextCursor":"c1"}`), &p); err != nil {
		t.Fatalf("legacy unmarshal error: %v", err)
	}
	if len(p.Items) != 1 || p.NextCursor != "c1" {
		t.Fatalf("unexpected legacy page: %+v", p)
	}

	var alt PageResult
	if err := json.Unmarshal([]byte(`{"records":[{"id":2}],"nextPage":"c2"}`), &alt); err != nil {
		t.Fatalf("alt unmarshal error: %v", err)
	}
	if len(alt.Items) != 1 || alt.NextCursor != "c2" {
		t.Fatalf("unexpected alt page: %+v", alt)
	}

	var bad PageResult
	if err := json.Unmarshal([]byte(`{`), &bad); err == nil {
		t.Fatalf("expected invalid json error")
	}
}

func TestQueryResultsDecodeError(t *testing.T) {
	q := QueryResults{{"bad": func() {}}}
	if err := q.Decode(&[]map[string]any{}); err == nil {
		t.Fatalf("expected decode error for unsupported value")
	}
}

func TestSchemaLookupHelpers(t *testing.T) {
	table := Table{
		Name:   "Users",
		Fields: []Field{{Name: "id"}, {Name: "email"}},
	}
	field, ok := table.Field("email")
	if !ok || field.Name != "email" {
		t.Fatalf("expected to find email field")
	}
	if _, ok := table.Field("missing"); ok {
		t.Fatalf("expected missing field to return false")
	}

	schema := Schema{Tables: []Table{table}}
	if _, ok := schema.Table("Users"); !ok {
		t.Fatalf("expected table lookup success")
	}
	if _, ok := schema.Table("Nope"); ok {
		t.Fatalf("expected table lookup failure")
	}
}

func TestSchemaJSONHelpersDefaults(t *testing.T) {
	if v := stringValue(123); v != "" {
		t.Fatalf("expected empty string for non-string input, got %q", v)
	}
	if v := boolValue("nope"); v {
		t.Fatalf("expected false for non-bool input")
	}
	if v := mapValue("x"); v != nil {
		t.Fatalf("expected nil map for non-map input")
	}
}

func TestSchemaFromEntitiesWithMetaAndResolvers(t *testing.T) {
	raw := []any{
		map[string]any{
			"name": "Thing",
			"identifier": map[string]any{
				"name": "id",
			},
			"attributes": []any{
				map[string]any{"name": "id", "type": "String"},
				map[string]any{"name": "label", "type": "String", "isNullable": true},
			},
			"resolvers": []any{
				"simple",
				map[string]any{"name": "complex", "resolver": "db.from(\"X\")", "meta": map[string]any{"a": 1}},
			},
			"indexes":  []any{map[string]any{"name": "idx_label"}},
			"triggers": []any{"t1", map[string]any{"name": "t2"}},
			"meta":     map[string]any{"tag": "x"},
		},
	}

	schema := schemaFromEntities(raw)
	if len(schema.Tables) != 1 {
		t.Fatalf("expected one table, got %d", len(schema.Tables))
	}
	tbl := schema.Tables[0]
	if len(tbl.Resolvers) != 2 || tbl.Resolvers[1].Resolver == "" || tbl.Resolvers[1].Meta["a"] != 1 {
		t.Fatalf("resolver parsing mismatch: %+v", tbl.Resolvers)
	}
	if len(tbl.Indexes) != 1 || tbl.Indexes[0].Name != "idx_label" {
		t.Fatalf("indexes not parsed: %+v", tbl.Indexes)
	}
	if len(tbl.Triggers) != 2 || tbl.Triggers[1] != "t2" {
		t.Fatalf("triggers not parsed: %+v", tbl.Triggers)
	}
	if tbl.Meta["tag"] != "x" {
		t.Fatalf("meta not preserved: %+v", tbl.Meta)
	}
}

func TestSchemaFromTablesArrayFullCoverage(t *testing.T) {
	raw := []any{
		map[string]any{
			"name": "Doc",
			"fields": []any{
				map[string]any{"name": "id", "type": "String", "nullable": false, "primaryKey": true},
				map[string]any{"name": "note", "type": "String", "nullable": true, "unique": true},
			},
			"resolvers": []any{
				"plain",
				map[string]any{"name": "withMeta", "resolver": "db.from(\"Doc\")", "meta": map[string]any{"x": "y"}},
			},
			"indexes":  []any{map[string]any{"name": "idx_note"}},
			"triggers": []any{map[string]any{"name": "t1"}, "t2"},
			"meta":     map[string]any{"owner": "team"},
		},
	}

	schema := schemaFromTablesArray(raw)
	if len(schema.Tables) != 1 {
		t.Fatalf("expected one table, got %d", len(schema.Tables))
	}
	tbl := schema.Tables[0]
	if !tbl.Fields[0].Primary || !tbl.Fields[1].Unique || !tbl.Fields[1].Nullable {
		t.Fatalf("field flags not parsed: %+v", tbl.Fields)
	}
	if len(tbl.Resolvers) != 2 || tbl.Resolvers[1].Meta["x"] != "y" {
		t.Fatalf("resolver meta missing: %+v", tbl.Resolvers)
	}
	if len(tbl.Triggers) != 2 || tbl.Triggers[0] != "t1" || tbl.Triggers[1] != "t2" {
		t.Fatalf("triggers parsed incorrectly: %+v", tbl.Triggers)
	}
	if tbl.Meta["owner"] != "team" {
		t.Fatalf("table meta not preserved")
	}
}

func TestParseSchemaJSONShapesAndErrors(t *testing.T) {
	raw := []byte(`{"schema":{"entities":[{"name":"X","attributes":[{"name":"id","type":"String"}]}]}}`)
	schema, err := ParseSchemaJSON(raw)
	if err != nil || len(schema.Tables) != 1 {
		t.Fatalf("expected nested schema parsed, got err=%v schema=%+v", err, schema)
	}

	invalid := []byte(`{`)
	if _, err := ParseSchemaJSON(invalid); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestPageResultUnmarshalCursorFallback(t *testing.T) {
	var p PageResult
	if err := json.Unmarshal([]byte(`{"records":[{"id":1}],"nextCursor":"cursor-only"}`), &p); err != nil {
		t.Fatalf("expected alt shape decode via nextCursor: %v", err)
	}
	if p.NextCursor != "cursor-only" || len(p.Items) != 1 {
		t.Fatalf("unexpected page result: %+v", p)
	}
}

func TestParseSchemaJSONEmptySchema(t *testing.T) {
	schema, err := ParseSchemaJSON([]byte(`{"schema":{}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Tables) != 0 {
		t.Fatalf("expected empty schema, got %+v", schema)
	}
}

func TestSchemaFromEntitiesSkipsInvalidEntries(t *testing.T) {
	schema := schemaFromEntities([]any{
		123,
		map[string]any{
			"name":       "Broken",
			"identifier": map[string]any{"name": "id"},
			"attributes": []any{"not-a-map"},
			"resolvers":  []any{123},
			"indexes":    []any{123},
			"triggers":   []any{123},
		},
	})
	if len(schema.Tables) != 1 || len(schema.Tables[0].Fields) != 0 {
		t.Fatalf("expected skip of invalid entries, got %+v", schema.Tables)
	}
	if len(schema.Tables[0].Resolvers) != 0 || len(schema.Tables[0].Indexes) != 0 || len(schema.Tables[0].Triggers) != 0 {
		t.Fatalf("expected no resolvers/indexes/triggers parsed from invalid types, got %+v", schema.Tables[0])
	}
}

func TestSchemaFromTablesArraySkipsInvalidEntries(t *testing.T) {
	schema := schemaFromTablesArray([]any{
		123,
		map[string]any{
			"name":      "Docs",
			"fields":    []any{123},
			"resolvers": []any{123},
			"indexes":   []any{123},
			"triggers":  []any{123},
		},
	})
	if len(schema.Tables) != 1 {
		t.Fatalf("expected table parsed despite invalid siblings")
	}
	tbl := schema.Tables[0]
	if len(tbl.Fields) != 0 {
		t.Fatalf("expected invalid fields skipped, got %+v", tbl.Fields)
	}
	if len(tbl.Resolvers) != 0 || len(tbl.Indexes) != 0 || len(tbl.Triggers) != 0 {
		t.Fatalf("expected invalid resolver/index/trigger entries skipped, got %+v", tbl)
	}
}
