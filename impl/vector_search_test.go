package impl

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestVectorSearchBuilderWire(t *testing.T) {
	query, err := contract.NewVectorSearchQuery(contract.VectorSearchQueryInput{
		Text:          "storm warning",
		MaxCandidates: 250,
	})
	if err != nil {
		t.Fatalf("new vector query: %v", err)
	}

	payload, err := json.Marshal(
		newQuery(nil, "Article").
			SearchVector(query).
			And(contract.Eq("published", true)),
	)
	if err != nil {
		t.Fatalf("marshal vector query builder: %v", err)
	}
	want := `{"type":"SelectQuery","table":"Article","conditions":{"conditionType":"CompoundCondition","conditions":[{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"MATCHES","value":{"maxCandidates":250,"minScore":null,"nearbyBucketRadius":1,"requireAllTerms":true,"semantic":null,"text":"storm warning"}}},{"conditionType":"SingleCondition","criteria":{"field":"published","operator":"EQUAL","value":true}}],"operator":"AND"}}`
	if string(payload) != want {
		t.Fatalf("unexpected vector builder wire:\n got: %s\nwant: %s", payload, want)
	}
}

func TestBoundedCandidateBuilderWire(t *testing.T) {
	lexical, err := contract.NewVectorSearchQuery(contract.VectorSearchQueryInput{
		Text:          "bounded lexical",
		MaxCandidates: 128,
	})
	if err != nil {
		t.Fatalf("new lexical query: %v", err)
	}
	assertQueryJSON(
		t,
		newQuery(nil, "Article").ApproximateSearch(lexical),
		`{"type":"SelectQuery","table":"Article","conditions":{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"SEARCH_CANDIDATES","value":{"maxCandidates":128,"minScore":null,"nearbyBucketRadius":1,"requireAllTerms":true,"semantic":null,"text":"bounded lexical"}}}}`,
	)

	hnsw, err := contract.NewHNSWSearchQuery(contract.HNSWSearchQueryInput{
		CalibrationID: -7909761245221418085,
		Vector:        []float64{0.25, -0.5, 0.75},
		MaxCandidates: 40,
		EFSearch:      96,
	})
	if err != nil {
		t.Fatalf("new HNSW query: %v", err)
	}
	assertQueryJSON(
		t,
		newQuery(nil, "Article").HNSWCandidates(hnsw),
		`{"type":"SelectQuery","table":"Article","conditions":{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"HNSW_CANDIDATES","value":{"calibrationId":"-7909761245221418085","efSearch":96,"formatVersion":1,"maxCandidates":40,"minScore":null,"vector":[0.25,-0.5,0.75]}}}}`,
	)

	assertQueryJSON(
		t,
		newQuery(nil, "Article").ApproximateCandidates("corpusId", []string{"one", "two"}, 17),
		`{"type":"SelectQuery","table":"Article","conditions":{"conditionType":"SingleCondition","criteria":{"field":"corpusId","operator":"CANDIDATES","value":{"maxCandidates":17,"values":["one","two"]}}}}`,
	)
}

func TestCandidateBuilderEnforcesSoleRootBothDirections(t *testing.T) {
	lexical, err := contract.NewVectorSearchQuery(contract.VectorSearchQueryInput{Text: "bounded"})
	if err != nil {
		t.Fatalf("new lexical query: %v", err)
	}

	tests := []struct {
		name  string
		query contract.Query
	}{
		{
			name: "candidate after existing criterion",
			query: newQuery(nil, "Article").
				Where(contract.Eq("published", true)).
				ApproximateSearch(lexical),
		},
		{
			name: "criterion after candidate",
			query: newQuery(nil, "Article").
				ApproximateSearch(lexical).
				And(contract.Eq("published", true)),
		},
		{
			name: "OR after candidate",
			query: newQuery(nil, "Article").
				ApproximateSearch(lexical).
				Or(contract.Eq("published", true)),
		},
		{
			name: "vector search after candidate",
			query: newQuery(nil, "Article").
				ApproximateSearch(lexical).
				SearchVector(lexical),
		},
		{
			name: "condition helper receives same enforcement",
			query: newQuery(nil, "Article").
				Where(contract.ApproximateCandidates("corpusId", "one")).
				And(contract.Eq("published", true)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := json.Marshal(test.query); err == nil || !strings.Contains(err.Error(), "sole root criterion") {
				t.Fatalf("expected sole-root error, got %v", err)
			}
		})
	}

	// The immutable source query remains valid when a derived composition fails.
	root := newQuery(nil, "Article").ApproximateSearch(lexical)
	_ = root.And(contract.Eq("published", true))
	if _, err := json.Marshal(root); err != nil {
		t.Fatalf("candidate root mutated by derived query: %v", err)
	}
}

func TestCandidateBuilderIsReadOnly(t *testing.T) {
	lexical, err := contract.NewVectorSearchQuery(contract.VectorSearchQueryInput{Text: "bounded"})
	if err != nil {
		t.Fatalf("new lexical query: %v", err)
	}
	query := newQuery(nil, "Article").ApproximateSearch(lexical)
	if _, err := query.Update(context.Background()); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("expected update read-only error, got %v", err)
	}
	if _, err := query.Delete(context.Background()); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("expected delete read-only error, got %v", err)
	}
}

func TestRawCandidateConditionsAreRecursivelyReadOnlyAndSoleRoot(t *testing.T) {
	tests := []struct {
		operator string
		field    string
	}{
		{operator: "CANDIDATES", field: "corpusId"},
		{operator: "SEARCH_CANDIDATES", field: "__full_text__"},
		{operator: "HNSW_CANDIDATES", field: "__full_text__"},
	}

	for _, test := range tests {
		t.Run(test.operator, func(t *testing.T) {
			root := rawOperatorCondition(test.field, test.operator)
			rootQuery := newQuery(nil, "Article").Where(root)
			assertReadOnlyMutation(t, rootQuery, test.operator)

			for name, query := range map[string]contract.Query{
				"candidate after criterion": newQuery(nil, "Article").
					Where(contract.Eq("active", true)).
					And(root),
				"criterion after candidate": newQuery(nil, "Article").
					Where(root).
					And(contract.Eq("active", true)),
				"candidate hidden in compound": newQuery(nil, "Article").Where(
					json.RawMessage(`{
						"conditionType":"CompoundCondition",
						"operator":"AND",
						"conditions":[
							{"conditionType":"SingleCondition","criteria":{"field":"active","operator":"EQUAL","value":true}},
							` + string(root) + `
						]
					}`),
				),
			} {
				t.Run(name, func(t *testing.T) {
					if _, err := json.Marshal(query); err == nil ||
						!strings.Contains(err.Error(), test.operator+" must be the sole root criterion") {
						t.Fatalf("expected recursive sole-root error, got %v", err)
					}
					assertReadOnlyMutation(t, query, test.operator)
				})
			}
		})
	}
}

func rawOperatorCondition(field, operator string) json.RawMessage {
	return json.RawMessage(`{"conditionType":"SingleCondition","criteria":{"field":"` +
		field + `","operator":"` + operator + `","value":{}}}`)
}

func assertReadOnlyMutation(t *testing.T, query contract.Query, operator string) {
	t.Helper()
	if _, err := query.Update(context.Background()); err == nil ||
		!strings.Contains(err.Error(), operator+" is read-only") {
		t.Fatalf("expected update read-only error for %s, got %v", operator, err)
	}
	if _, err := query.Delete(context.Background()); err == nil ||
		!strings.Contains(err.Error(), operator+" is read-only") {
		t.Fatalf("expected delete read-only error for %s, got %v", operator, err)
	}
}

func TestQueryMarshalPropagatesNativeSearchValidation(t *testing.T) {
	invalid := contract.HNSWSearchQuery{
		CalibrationID: "1",
		Vector:        []float64{0},
		MaxCandidates: 1,
		EFSearch:      1,
		FormatVersion: contract.HNSWQueryFormatVersion,
	}
	query := newQuery(nil, "Article").HNSWCandidates(invalid)
	if _, err := json.Marshal(query); err == nil || !strings.Contains(err.Error(), "non-zero finite norm") {
		t.Fatalf("expected HNSW validation error, got %v", err)
	}
}

func assertQueryJSON(t *testing.T, query contract.Query, want string) {
	t.Helper()
	payload, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("marshal query: %v", err)
	}
	if string(payload) != want {
		t.Fatalf("unexpected query JSON:\n got: %s\nwant: %s", payload, want)
	}
}
