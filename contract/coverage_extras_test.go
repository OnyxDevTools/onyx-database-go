package contract

import "testing"

func TestParseSchemaJSONNestedAndTables(t *testing.T) {
	raw := []byte(`{"schema":{"tables":[{"name":"Thing","fields":[{"name":"id","type":"String","primaryKey":true}]}]}}`)
	s, err := ParseSchemaJSON(raw)
	if err != nil {
		t.Fatalf("parse nested schema: %v", err)
	}
	if len(s.Tables) != 1 || s.Tables[0].Name != "Thing" {
		t.Fatalf("unexpected schema: %+v", s)
	}
}

func TestNormalizeSchemaSortingAcrossTables(t *testing.T) {
	s := Schema{
		Tables: []Table{
			{
				Name: "Z",
				Fields: []Field{
					{Name: "b"}, {Name: "a"},
				},
				Resolvers: []Resolver{{Name: "beta"}, {Name: "alpha"}},
				Indexes:   []Index{{Name: "z"}, {Name: "a"}},
				Triggers:  []string{"t2", "t1"},
			},
			{Name: "A"},
		},
	}
	n := NormalizeSchema(s)
	if n.Tables[0].Name != "A" || n.Tables[1].Fields[0].Name != "a" || n.Tables[1].Resolvers[0].Name != "alpha" {
		t.Fatalf("normalization failed: %+v", n)
	}
}

func TestHelperValueExtractors(t *testing.T) {
	if stringValue(123) != "" {
		t.Fatalf("expected empty string")
	}
	if boolValue("nope") {
		t.Fatalf("expected false")
	}
	if mapValue("x") != nil {
		t.Fatalf("expected nil map")
	}
}

