package contract

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// SearchMode selects how searchable text is matched.
type SearchMode string

const (
	// SearchModeLexical matches indexed terms.
	SearchModeLexical SearchMode = "lexical"
	// SearchModeSemantic embeds the query text and uses semantic similarity.
	SearchModeSemantic SearchMode = "semantic"
	// SearchModeHybrid combines lexical and semantic matching.
	SearchModeHybrid SearchMode = "hybrid"
)

// SearchMatch controls how the lexical portion of a search treats query terms.
type SearchMatch string

const (
	// SearchMatchAny allows records matching any normalized query term.
	SearchMatchAny SearchMatch = "any"
	// SearchMatchAll requires records to match every normalized query term.
	SearchMatchAll SearchMatch = "all"
)

// SearchOptions configures a high-level lexical, semantic, or hybrid search.
// Its zero value selects hybrid search, matches any lexical term, and uses a
// candidate budget of 1000 with no minimum score.
type SearchOptions struct {
	Mode          SearchMode  `json:"mode"`
	Match         SearchMatch `json:"match,omitempty"`
	MinScore      *float64    `json:"minScore,omitempty"`
	MaxCandidates int         `json:"maxCandidates,omitempty"`
}

type textSearchQuery struct {
	Text          string      `json:"text"`
	Mode          SearchMode  `json:"mode"`
	Match         SearchMatch `json:"match"`
	MinScore      *float64    `json:"minScore"`
	MaxCandidates int         `json:"maxCandidates"`
}

func newTextSearchQuery(text string, options SearchOptions) textSearchQuery {
	mode := options.Mode
	if mode == "" {
		mode = SearchModeHybrid
	}
	match := options.Match
	if match == "" {
		match = SearchMatchAny
	}
	maxCandidates := options.MaxCandidates
	if maxCandidates == 0 {
		maxCandidates = DefaultVectorSearchCandidates
	}
	return textSearchQuery{
		Text:          text,
		Mode:          mode,
		Match:         match,
		MinScore:      copyFloat64(options.MinScore),
		MaxCandidates: maxCandidates,
	}
}

func (q textSearchQuery) MarshalJSON() ([]byte, error) {
	if strings.TrimSpace(q.Text) == "" {
		return nil, fmt.Errorf("search text must not be blank")
	}
	switch q.Mode {
	case SearchModeLexical, SearchModeSemantic, SearchModeHybrid:
	default:
		return nil, fmt.Errorf("search mode must be lexical, semantic, or hybrid")
	}
	switch q.Match {
	case SearchMatchAny, SearchMatchAll:
	default:
		return nil, fmt.Errorf("search match must be any or all")
	}
	if q.MinScore != nil && (math.IsNaN(*q.MinScore) || math.IsInf(*q.MinScore, 0) || *q.MinScore < 0 || *q.MinScore > 1) {
		return nil, fmt.Errorf("search minScore must be finite and between 0 and 1")
	}
	minimumCandidates := 1
	if q.Mode == SearchModeHybrid {
		minimumCandidates = 2
	}
	if q.MaxCandidates < minimumCandidates || q.MaxCandidates > MaxVectorSearchCandidates {
		return nil, fmt.Errorf(
			"search maxCandidates must be between %d and %d for %s mode",
			minimumCandidates,
			MaxVectorSearchCandidates,
			q.Mode,
		)
	}

	type wire textSearchQuery
	return json.Marshal(wire(q))
}
