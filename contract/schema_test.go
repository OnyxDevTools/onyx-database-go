package contract

import (
	"encoding/json"
	"testing"
)

func TestParseAndNormalizeSchema(t *testing.T) {
	raw := []byte(`{
        "tables": [
            {"name": "Order", "fields": [
                {"name": "total", "type": "float"},
                {"name": "id", "type": "string", "primaryKey": true}
            ]},
            {"name": "User", "fields": [
                {"name": "email", "type": "string", "unique": true},
                {"name": "id", "type": "string", "primaryKey": true}
            ]}
        ]
    }`)

	schema, err := ParseSchemaJSON(raw)
	if err != nil {
		t.Fatalf("parse schema: %v", err)
	}

	normalized := NormalizeSchema(schema)

	if len(normalized.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(normalized.Tables))
	}

	if normalized.Tables[0].Name != "Order" || normalized.Tables[1].Name != "User" {
		t.Fatalf("tables not sorted: %+v", normalized.Tables)
	}

	order := normalized.Tables[0]
	if order.Fields[0].Name != "id" || order.Fields[1].Name != "total" {
		t.Fatalf("fields not sorted for Order: %+v", order.Fields)
	}

	user, ok := normalized.Table("User")
	if !ok {
		t.Fatalf("missing User table")
	}

	if _, ok := user.Field("email"); !ok {
		t.Fatalf("email field lookup failed")
	}

	// Ensure JSON round-trips deterministically after normalization.
	normalizedJSON, err := json.Marshal(normalized)
	if err != nil {
		t.Fatalf("marshal normalized: %v", err)
	}

	expected := `{"tables":[{"name":"Order","fields":[{"name":"id","type":"string","primaryKey":true},{"name":"total","type":"float"}]},{"name":"User","fields":[{"name":"email","type":"string","unique":true},{"name":"id","type":"string","primaryKey":true}]}]}`
	if string(normalizedJSON) != expected {
		t.Fatalf("unexpected normalized json: %s", string(normalizedJSON))
	}
}
