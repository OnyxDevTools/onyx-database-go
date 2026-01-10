package schema

import (
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestDiffSchemasPassthrough(t *testing.T) {
	base := contract.Schema{Tables: []contract.Table{{Name: "base", Fields: []contract.Field{{Name: "id", Type: "string"}}}}}
	updated := contract.Schema{Tables: []contract.Table{{Name: "updated", Fields: []contract.Field{{Name: "id", Type: "string"}}}}}

	diff := DiffSchemas(base, updated)
	if len(diff.AddedTables) != 1 || diff.AddedTables[0].Name != "updated" {
		t.Fatalf("expected added table, got %+v", diff.AddedTables)
	}
	if len(diff.RemovedTables) != 1 || diff.RemovedTables[0].Name != "base" {
		t.Fatalf("expected removed table, got %+v", diff.RemovedTables)
	}
}
