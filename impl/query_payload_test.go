package impl

import (
	"encoding/json"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestBuildUpdatePayloadIncludesUpdatesAndSort(t *testing.T) {
	q := &query{
		table: "users",
		updates: map[string]any{
			"email": "a@example.com",
		},
		sorts: []contract.Sort{contract.Asc("email")},
	}
	payload, err := buildUpdatePayload(q)
	if err != nil {
		t.Fatalf("build update payload: %v", err)
	}
	if payload.Table != "users" {
		t.Fatalf("unexpected table: %s", payload.Table)
	}
	if payload.Updates["email"] != "a@example.com" {
		t.Fatalf("missing updates: %+v", payload.Updates)
	}
	if len(payload.Sort) != 1 {
		t.Fatalf("expected sort")
	}
	if payload.Limit != nil {
		t.Fatalf("expected nil limit")
	}

	encoded, err := json.Marshal(payload)
	if err != nil || len(encoded) == 0 {
		t.Fatalf("marshal payload err: %v", err)
	}
}

func TestBuildQueryPayloadVariants(t *testing.T) {
	limit := 5
	q := &query{
		table:         "users",
		selectFields:  []string{"id"},
		groupFields:   []string{"status"},
		distinct:      true,
		resolveFields: []string{"roles"},
		sorts:         []contract.Sort{contract.Desc("createdAt")},
		limit:         &limit,
		clauses:       []clause{{Type: "and", Condition: contract.Eq("status", "active")}},
	}

	withLimit, err := buildQueryPayload(q, true)
	if err != nil {
		t.Fatalf("build query payload: %v", err)
	}
	if withLimit.Limit == nil || *withLimit.Limit != 5 {
		t.Fatalf("expected limit included")
	}
	if len(withLimit.Fields) != 1 || len(withLimit.GroupBy) != 1 || len(withLimit.Resolvers) != 1 || len(withLimit.Sort) != 1 {
		t.Fatalf("expected fields/group/resolvers/sort set: %+v", withLimit)
	}
	if withLimit.Distinct == nil || !*withLimit.Distinct {
		t.Fatalf("expected distinct included")
	}

	withoutLimit, err := buildQueryPayload(q, false)
	if err != nil {
		t.Fatalf("build query payload without limit: %v", err)
	}
	if withoutLimit.Limit != nil {
		t.Fatalf("expected limit omitted when includeLimit=false")
	}
	if withoutLimit.Conditions == nil {
		t.Fatalf("expected conditions built")
	}

	// No clauses case should return nil conditions
	if cond, err := buildConditions(nil); err != nil || cond != nil {
		t.Fatalf("expected nil conditions when no clauses")
	}

	part := "p1"
	q.partition = &part
	withPartition, err := buildQueryPayload(q, true)
	if err != nil {
		t.Fatalf("build partition query payload: %v", err)
	}
	if withPartition.Partition == nil || *withPartition.Partition != "p1" {
		t.Fatalf("expected partition to be set")
	}

	upd, err := buildUpdatePayload(q)
	if err != nil {
		t.Fatalf("build partition update payload: %v", err)
	}
	if upd.Partition == nil || *upd.Partition != "p1" {
		t.Fatalf("expected partition on update payload")
	}
}
