package contract

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestSearchWithOptionsCanonicalWire(t *testing.T) {
	condition := SearchWithOptions(
		"how do i calculate cost per horse",
		SearchOptions{},
	)
	data, err := json.Marshal(condition)
	if err != nil {
		t.Fatalf("marshal high-level search: %v", err)
	}
	want := `{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"SEARCH","value":{"text":"how do i calculate cost per horse","mode":"hybrid","match":"any","minScore":null,"maxCandidates":1000}}}`
	if string(data) != want {
		t.Fatalf("unexpected search wire:\n got: %s\nwant: %s", data, want)
	}

	minScore := 0.4
	explicit := SearchWithOptions("cost per horse", SearchOptions{
		Mode:          SearchModeLexical,
		Match:         SearchMatchAll,
		MinScore:      &minScore,
		MaxCandidates: 500,
	})
	explicitData, err := json.Marshal(explicit)
	if err != nil {
		t.Fatalf("marshal explicit search: %v", err)
	}
	explicitWant := `{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"SEARCH","value":{"text":"cost per horse","mode":"lexical","match":"all","minScore":0.4,"maxCandidates":500}}}`
	if string(explicitData) != explicitWant {
		t.Fatalf("unexpected explicit search wire:\n got: %s\nwant: %s", explicitData, explicitWant)
	}
}

func TestSearchWithOptionsAcceptsEveryMode(t *testing.T) {
	for _, mode := range []SearchMode{SearchModeLexical, SearchModeSemantic, SearchModeHybrid} {
		t.Run(string(mode), func(t *testing.T) {
			if _, err := json.Marshal(SearchWithOptions("query", SearchOptions{Mode: mode})); err != nil {
				t.Fatalf("mode %q should be valid: %v", mode, err)
			}
		})
	}
}

func TestSearchWithOptionsValidation(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		options  SearchOptions
		contains string
	}{
		{name: "blank text", text: "  ", options: SearchOptions{Mode: SearchModeHybrid}, contains: "must not be blank"},
		{name: "invalid mode", text: "query", options: SearchOptions{Mode: "both"}, contains: "mode"},
		{name: "invalid match", text: "query", options: SearchOptions{Mode: SearchModeLexical, Match: "some"}, contains: "match"},
		{name: "nan score", text: "query", options: SearchOptions{Mode: SearchModeSemantic, MinScore: searchFloatPointer(math.NaN())}, contains: "minScore"},
		{name: "infinite score", text: "query", options: SearchOptions{Mode: SearchModeSemantic, MinScore: searchFloatPointer(math.Inf(1))}, contains: "minScore"},
		{name: "negative score", text: "query", options: SearchOptions{Mode: SearchModeSemantic, MinScore: searchFloatPointer(-0.01)}, contains: "minScore"},
		{name: "score above one", text: "query", options: SearchOptions{Mode: SearchModeSemantic, MinScore: searchFloatPointer(1.01)}, contains: "minScore"},
		{name: "negative candidates", text: "query", options: SearchOptions{Mode: SearchModeHybrid, MaxCandidates: -1}, contains: "maxCandidates"},
		{name: "hybrid needs two candidates", text: "query", options: SearchOptions{Mode: SearchModeHybrid, MaxCandidates: 1}, contains: "maxCandidates"},
		{name: "too many candidates", text: "query", options: SearchOptions{Mode: SearchModeHybrid, MaxCandidates: MaxVectorSearchCandidates + 1}, contains: "maxCandidates"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := json.Marshal(SearchWithOptions(test.text, test.options))
			if err == nil || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("expected error containing %q, got %v", test.contains, err)
			}
		})
	}
}

func TestSearchWithOptionsCandidateMinimumDependsOnMode(t *testing.T) {
	for _, mode := range []SearchMode{SearchModeLexical, SearchModeSemantic} {
		t.Run(string(mode), func(t *testing.T) {
			if _, err := json.Marshal(SearchWithOptions("query", SearchOptions{
				Mode:          mode,
				MaxCandidates: 1,
			})); err != nil {
				t.Fatalf("mode %q should accept one candidate: %v", mode, err)
			}
		})
	}
}

func TestSearchWithOptionsMinScoreBoundaries(t *testing.T) {
	for _, score := range []float64{0, 1} {
		if _, err := json.Marshal(SearchWithOptions("query", SearchOptions{
			MinScore: &score,
		})); err != nil {
			t.Fatalf("minScore %v should be valid: %v", score, err)
		}
	}
}

func TestSearchWithOptionsCopiesMinScore(t *testing.T) {
	score := 0.25
	condition := SearchWithOptions("query", SearchOptions{
		Mode:     SearchModeSemantic,
		MinScore: &score,
	})
	score = math.NaN()
	if _, err := json.Marshal(condition); err != nil {
		t.Fatalf("search condition retained caller-owned score pointer: %v", err)
	}
}

func searchFloatPointer(value float64) *float64 { return &value }
