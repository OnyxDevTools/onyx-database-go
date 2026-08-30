package impl

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
)

func TestSearchWithOptionsBuilderWireAndComposition(t *testing.T) {
	options := contract.SearchOptions{
		Mode:          contract.SearchModeHybrid,
		Match:         contract.SearchMatchAny,
		MaxCandidates: 500,
	}
	payload, err := json.Marshal(
		newQuery(nil, "ActiveDocumentChunk").
			SearchWithOptions("how do i calculate cost per horse", options).
			And(contract.Eq("active", true)),
	)
	if err != nil {
		t.Fatalf("marshal composed search: %v", err)
	}
	want := `{"type":"SelectQuery","table":"ActiveDocumentChunk","conditions":{"conditionType":"CompoundCondition","conditions":[{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"SEARCH","value":{"match":"any","maxCandidates":500,"minScore":null,"mode":"hybrid","text":"how do i calculate cost per horse"}}},{"conditionType":"SingleCondition","criteria":{"field":"active","operator":"EQUAL","value":true}}],"operator":"AND"}}`
	if string(payload) != want {
		t.Fatalf("unexpected composed search payload:\n got: %s\nwant: %s", payload, want)
	}

	// SEARCH is composable in either direction and is not subject to the
	// sole-root restriction used by bounded candidate operators.
	reversed := newQuery(nil, "ActiveDocumentChunk").
		Where(contract.Eq("active", true)).
		SearchWithOptions("cost per horse", options)
	if _, err := json.Marshal(reversed); err != nil {
		t.Fatalf("marshal reversed composed search: %v", err)
	}
}

func TestSearchWithOptionsAllTablesWire(t *testing.T) {
	client := &client{cfg: resolver.ResolvedConfig{DatabaseID: "db", Partition: "tenant-default"}}
	payload, err := json.Marshal(client.SearchWithOptions(
		"horse expenses",
		contract.SearchOptions{Mode: contract.SearchModeSemantic},
	))
	if err != nil {
		t.Fatalf("marshal all-table search: %v", err)
	}
	want := `{"type":"SelectQuery","table":"ALL","conditions":{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"SEARCH","value":{"match":"any","maxCandidates":1000,"minScore":null,"mode":"semantic","text":"horse expenses"}}}}`
	if string(payload) != want {
		t.Fatalf("unexpected all-table search payload:\n got: %s\nwant: %s", payload, want)
	}

	legacy, err := json.Marshal(client.Search("horse expenses"))
	if err != nil {
		t.Fatalf("marshal legacy all-table search: %v", err)
	}
	if strings.Contains(string(legacy), `"partition"`) {
		t.Fatalf("legacy all-table search inherited the default partition: %s", legacy)
	}

	direct, err := json.Marshal(client.
		From("ActiveDocumentChunk").
		SearchWithOptions("horse expenses", contract.SearchOptions{Mode: contract.SearchModeSemantic}))
	if err != nil {
		t.Fatalf("marshal direct high-level search: %v", err)
	}
	if !strings.Contains(string(direct), `"partition":"tenant-default"`) {
		t.Fatalf("direct table search did not inherit the default partition: %s", direct)
	}
}

func TestSearchWithOptionsIsReadOnly(t *testing.T) {
	query := newQuery(nil, "Article").
		SearchWithOptions("storm warning", contract.SearchOptions{Mode: contract.SearchModeHybrid}).
		And(contract.Eq("published", true))

	if _, err := query.Update(context.Background()); err == nil || !strings.Contains(err.Error(), "SEARCH is read-only") {
		t.Fatalf("expected update read-only error, got %v", err)
	}
	if _, err := query.Delete(context.Background()); err == nil || !strings.Contains(err.Error(), "SEARCH is read-only") {
		t.Fatalf("expected delete read-only error, got %v", err)
	}
}

func TestNestedCustomSearchConditionIsReadOnly(t *testing.T) {
	nested := json.RawMessage(`{
		"conditionType":"CompoundCondition",
		"operator":"AND",
		"conditions":[
			{"conditionType":"SingleCondition","criteria":{"field":"active","operator":"EQUAL","value":true}},
			{"conditionType":"CompoundCondition","operator":"OR","conditions":[
				{"conditionType":"SingleCondition","criteria":{"field":"title","operator":"EQUAL","value":"guide"}},
				{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"SEARCH","value":{"text":"horse expenses","mode":"semantic","match":"any","minScore":null,"maxCandidates":1000}}}
			]}
		]
	}`)
	query := newQuery(nil, "Article").Where(nested)

	if _, err := query.Update(context.Background()); err == nil || !strings.Contains(err.Error(), "SEARCH is read-only") {
		t.Fatalf("expected nested update read-only error, got %v", err)
	}
	if _, err := query.Delete(context.Background()); err == nil || !strings.Contains(err.Error(), "SEARCH is read-only") {
		t.Fatalf("expected nested delete read-only error, got %v", err)
	}
}

func TestNestedReadOnlyDetectionIgnoresSearchTextInsideOrdinaryValues(t *testing.T) {
	raw := json.RawMessage(`{
		"conditionType":"SingleCondition",
		"criteria":{"field":"metadata","operator":"EQUAL","value":{"operator":"SEARCH"}}
	}`)
	if operator := nestedReadOnlyOperator(raw); operator != "" {
		t.Fatalf("ordinary condition value was mistaken for %s", operator)
	}
}

func TestSearchWithOptionsRejectsConflictingFullTextCriteria(t *testing.T) {
	options := contract.SearchOptions{Mode: contract.SearchModeLexical}
	vector, err := contract.NewVectorSearchQuery(contract.VectorSearchQueryInput{Text: "exact"})
	if err != nil {
		t.Fatalf("new vector query: %v", err)
	}

	tests := []struct {
		name     string
		query    contract.Query
		contains string
	}{
		{
			name: "two high-level searches",
			query: newQuery(nil, "Article").
				SearchWithOptions("one", options).
				SearchWithOptions("two", options),
			contains: "only one SEARCH criterion",
		},
		{
			name: "legacy search after high-level search",
			query: newQuery(nil, "Article").
				SearchWithOptions("semantic", options).
				Search("legacy"),
			contains: "another full-text search criterion",
		},
		{
			name: "high-level search after legacy search",
			query: newQuery(nil, "Article").
				Search("legacy").
				SearchWithOptions("semantic", options),
			contains: "another full-text search criterion",
		},
		{
			name: "native vector search after high-level search",
			query: newQuery(nil, "Article").
				SearchWithOptions("semantic", options).
				SearchVector(vector),
			contains: "another full-text search criterion",
		},
		{
			name: "candidate conflicts with high-level search",
			query: newQuery(nil, "Article").
				SearchWithOptions("semantic", options).
				ApproximateCandidates("corpusId", "one"),
			contains: "CANDIDATES must be the sole root criterion",
		},
		{
			name: "raw compound full-text conflict",
			query: newQuery(nil, "Article").Where(json.RawMessage(`{
				"conditionType":"CompoundCondition",
				"operator":"AND",
				"conditions":[
					{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"SEARCH","value":{"text":"semantic","mode":"semantic","match":"any","minScore":null,"maxCandidates":1000}}},
					{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"MATCHES","value":{"queryText":"legacy","minScore":null}}}
				]
			}`)),
			contains: "another full-text search criterion",
		},
		{
			name: "raw search requires full-text field",
			query: newQuery(nil, "Article").Where(json.RawMessage(`{
				"conditionType":"SingleCondition",
				"criteria":{"field":"title","operator":"SEARCH","value":{"text":"semantic","mode":"semantic","match":"any","minScore":null,"maxCandidates":1000}}
			}`)),
			contains: "SEARCH requires the __full_text__ field",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := json.Marshal(test.query); err == nil || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("expected %q error, got %v", test.contains, err)
			}
		})
	}
}
