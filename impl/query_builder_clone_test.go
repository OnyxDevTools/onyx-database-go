package impl

import (
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestQueryCloneCopiesSlicesAndMaps(t *testing.T) {
	limit := 10
	orig := &query{
		table:         "users",
		selectFields:  []string{"a"},
		groupFields:   []string{"g"},
		resolveFields: []string{"r"},
		sorts:         []contract.Sort{contract.Asc("a")},
		limit:         &limit,
		updates:       map[string]any{"name": "a"},
		clauses:       []clause{{Type: "and", Condition: contract.Eq("id", 1)}},
		partition:     ptr("p1"),
	}

	clone := orig.clone()
	if clone == orig || &clone.selectFields[0] == &orig.selectFields[0] {
		t.Fatalf("expected deep copy of slices")
	}

	orig.selectFields[0] = "changed"
	orig.updates["name"] = "b"
	if clone.selectFields[0] != "a" || clone.updates["name"] != "a" {
		t.Fatalf("clone mutated with original changes: %+v", clone)
	}
	if clone.partition == nil || *clone.partition != "p1" {
		t.Fatalf("expected partition copied, got %+v", clone.partition)
	}
}

func ptr[T any](v T) *T { return &v }
