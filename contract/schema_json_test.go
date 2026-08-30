package contract

import "testing"

func TestParseSchemaJSONTables(t *testing.T) {
	data := []byte(`{
		"tables": [
			{"name": "users", "fields": [
				{"name": "id", "type": "string", "primaryKey": true, "nullable": false},
				{"name": "email", "type": "string", "nullable": true}
			],
			"resolvers": [{"name":"profile"}]}
		]
	}`)

	schema, err := ParseSchemaJSON(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(schema.Tables) != 1 || schema.Tables[0].Name != "users" {
		t.Fatalf("unexpected tables: %+v", schema.Tables)
	}

	id := schema.Tables[0].Fields[0]
	if !id.Primary || id.Nullable {
		t.Fatalf("expected id to be primary and non-nullable, got %+v", id)
	}

	if len(schema.Tables[0].Resolvers) != 1 || schema.Tables[0].Resolvers[0].Name != "profile" {
		t.Fatalf("expected resolver names, got %+v", schema.Tables[0].Resolvers)
	}
}

func TestParseSchemaJSONEntities(t *testing.T) {
	data := []byte(`{
		"entities": [
			{
				"name": "User",
				"type": "SEARCHABLE",
				"searchSupport": "SEMANTIC",
				"partition": "tenantId",
				"identifier": {"name": "id", "type": "String", "generator": "UUID"},
				"attributes": [
					{"name": "id", "type": "String", "isNullable": false},
					{"name": "email", "type": "String", "isNullable": true}
				],
				"resolvers": [{"name":"roles"}, {"name":"profile"}],
				"indexes": [
					{"name":"email","type":"DEFAULT"},
					{"name":"content","type":"VECTOR","minimumScore":0.5}
				]
			}
		]
	}`)

	schema, err := ParseSchemaJSON(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(schema.Tables) != 1 || schema.Tables[0].Name != "User" {
		t.Fatalf("unexpected tables: %+v", schema.Tables)
	}
	if schema.Tables[0].Type != TableTypeSearchable || schema.Tables[0].SearchSupport != SearchSupportSemantic || schema.Tables[0].Partition != "tenantId" {
		t.Fatalf("expected searchable partitioned table metadata, got %+v", schema.Tables[0])
	}
	if len(schema.Tables[0].Indexes) != 2 || schema.Tables[0].Indexes[1].Type != IndexTypeVector {
		t.Fatalf("expected native index metadata, got %+v", schema.Tables[0].Indexes)
	}
	if score := schema.Tables[0].Indexes[1].MinimumScore; score == nil || *score != 0.5 {
		t.Fatalf("expected legacy minimumScore preserved, got %+v", score)
	}

	if len(schema.Tables[0].Fields) != 2 {
		t.Fatalf("expected 2 fields, got %+v", schema.Tables[0].Fields)
	}

	var id Field
	for _, f := range schema.Tables[0].Fields {
		if f.Name == "id" {
			id = f
			break
		}
	}
	if !id.Primary || id.Type != "String" {
		t.Fatalf("expected primary id of type String, got %+v", id)
	}

	if len(schema.Tables[0].Resolvers) != 2 || schema.Tables[0].Resolvers[0].Name == "" {
		t.Fatalf("expected resolver names, got %+v", schema.Tables[0].Resolvers)
	}
}

func TestParseSchemaJSONNestedSchemaObject(t *testing.T) {
	data := []byte(`{
		"schema": {
			"tables": [
				{"name": "roles", "fields": [{"name": "id", "type": "string"}]}
			]
		}
	}`)

	schema, err := ParseSchemaJSON(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(schema.Tables) != 1 || schema.Tables[0].Name != "roles" {
		t.Fatalf("unexpected tables: %+v", schema.Tables)
	}
}
