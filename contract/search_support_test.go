package contract

import (
	"encoding/json"
	"testing"
)

func TestTableSearchSupportJSON(t *testing.T) {
	table := Table{
		Name:          "Article",
		Type:          TableTypeSearchable,
		SearchSupport: SearchSupportBoth,
		Fields:        []Field{},
	}

	encoded, err := json.Marshal(table)
	if err != nil {
		t.Fatalf("marshal table: %v", err)
	}
	if string(encoded) != `{"name":"Article","type":"SEARCHABLE","searchSupport":"BOTH","fields":[]}` {
		t.Fatalf("unexpected table JSON: %s", encoded)
	}
}
