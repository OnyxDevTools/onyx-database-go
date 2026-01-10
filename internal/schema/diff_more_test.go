package schema

import (
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestDiffSchemasSortingAndComparatorPaths(t *testing.T) {
	base := contract.Schema{
		Tables: []contract.Table{
			{
				Name: "base",
				Fields: []contract.Field{
					{Name: "id", Type: "string"},
					{Name: "f1", Type: "string"},
					{Name: "f2", Type: "string"},
					{Name: "old1", Type: "string"},
					{Name: "old2", Type: "string"},
				},
				Resolvers: []contract.Resolver{
					{Name: "baseOnly1"},
					{Name: "baseOnly2"},
					{Name: "modR1", Resolver: "sql1", Meta: map[string]any{"k": "v"}},
					{Name: "modR2", Resolver: "sql2"},
				},
			},
			{
				Name:   "removedA",
				Fields: []contract.Field{{Name: "id", Type: "string"}},
			},
			{
				Name:   "removedB",
				Fields: []contract.Field{{Name: "id", Type: "string"}},
			},
			{
				Name:   "other",
				Fields: []contract.Field{{Name: "old", Type: "string"}},
			},
		},
	}

	updated := contract.Schema{
		Tables: []contract.Table{
			{
				Name: "base",
				Fields: []contract.Field{
					{Name: "id", Type: "string", Nullable: true}, // modified
					{Name: "f1", Type: "int"},                    // modified
					{Name: "f2", Type: "string", Nullable: true}, // modified
					{Name: "new1", Type: "string"},               // added
					{Name: "new2", Type: "string"},               // added
				},
				Resolvers: []contract.Resolver{
					{Name: "newOnly1"},
					{Name: "newOnly2"},
					{Name: "modR1", Resolver: "sql1-new", Meta: map[string]any{"k": "v"}},
					{Name: "modR2", Resolver: "sql2", Meta: map[string]any{"extra": "y"}},
				},
			},
			{
				Name:   "addedB",
				Fields: []contract.Field{{Name: "id", Type: "string"}},
			},
			{
				Name:   "addedA",
				Fields: []contract.Field{{Name: "id", Type: "string"}},
			},
			{
				Name: "other",
				Fields: []contract.Field{
					{Name: "old", Type: "int"},    // modified
					{Name: "new", Type: "string"}, // added
				},
			},
		},
	}

	diff := DiffSchemas(base, updated)

	if got := []string{diff.AddedTables[0].Name, diff.AddedTables[1].Name}; !(got[0] == "addedA" && got[1] == "addedB") {
		t.Fatalf("expected added tables sorted, got %v", got)
	}
	if got := []string{diff.RemovedTables[0].Name, diff.RemovedTables[1].Name}; !(got[0] == "removedA" && got[1] == "removedB") {
		t.Fatalf("expected removed tables sorted, got %v", got)
	}
	if len(diff.TableDiffs) != 2 || diff.TableDiffs[0].Name != "base" || diff.TableDiffs[1].Name != "other" {
		t.Fatalf("expected sorted table diffs for base and other, got %+v", diff.TableDiffs)
	}

	baseDiff := diff.TableDiffs[0]
	if len(baseDiff.AddedFields) != 2 || baseDiff.AddedFields[0].Name != "new1" || baseDiff.AddedFields[1].Name != "new2" {
		t.Fatalf("expected sorted added fields, got %+v", baseDiff.AddedFields)
	}
	if len(baseDiff.RemovedFields) != 2 || baseDiff.RemovedFields[0].Name != "old1" || baseDiff.RemovedFields[1].Name != "old2" {
		t.Fatalf("expected sorted removed fields, got %+v", baseDiff.RemovedFields)
	}
	if len(baseDiff.ModifiedFields) < 2 {
		t.Fatalf("expected multiple modified fields, got %+v", baseDiff.ModifiedFields)
	}
	if len(baseDiff.AddedResolvers) != 2 || baseDiff.AddedResolvers[0] != "baseOnly1" || baseDiff.AddedResolvers[1] != "baseOnly2" {
		t.Fatalf("expected sorted added resolvers, got %+v", baseDiff.AddedResolvers)
	}
	if len(baseDiff.RemovedResolvers) != 2 || baseDiff.RemovedResolvers[0] != "newOnly1" || baseDiff.RemovedResolvers[1] != "newOnly2" {
		t.Fatalf("expected sorted removed resolvers, got %+v", baseDiff.RemovedResolvers)
	}
	if len(baseDiff.ModifiedResolvers) != 2 || baseDiff.ModifiedResolvers[0].Name != "modR1" || baseDiff.ModifiedResolvers[1].Name != "modR2" {
		t.Fatalf("expected sorted modified resolvers, got %+v", baseDiff.ModifiedResolvers)
	}
}
