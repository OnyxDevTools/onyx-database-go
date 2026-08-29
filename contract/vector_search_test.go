package contract

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestSemanticVectorSignatureCanonicalWire(t *testing.T) {
	signature, err := NewSemanticVectorSignature(
		-7909761245221418085,
		6,
		[]int{1, 2},
		[]int{3, 4},
		[]int64{0x0123456789abcdef},
		0.25,
	)
	if err != nil {
		t.Fatalf("new semantic signature: %v", err)
	}

	data, err := json.Marshal(signature)
	if err != nil {
		t.Fatalf("marshal semantic signature: %v", err)
	}
	want := `{"calibrationId":"-7909761245221418085","bucketId":6,"cells":[1,2],"cellCounts":[3,4],"fingerprint":["0x0123456789abcdef"],"bands":["0x000000000000cdef","0x00000000000089ab","0x0000000000004567","0x0000000000000123"],"boundaryConfidence":0.25}`
	if string(data) != want {
		t.Fatalf("unexpected semantic wire:\n got: %s\nwant: %s", data, want)
	}

	// Direct wire values are accepted in their lossless decimal/hex forms and
	// normalized to the same canonical lower-case, fixed-width output.
	direct := SemanticVectorSignature{
		CalibrationID:      " -7909761245221418085 ",
		BucketID:           6,
		Cells:              []int{1, 2},
		CellCounts:         []int{3, 4},
		Fingerprint:        []string{"81985529216486895"},
		Bands:              []string{"0xCDEF", "0x89AB", "0x4567", "0x0123"},
		BoundaryConfidence: 0.25,
	}
	directData, err := json.Marshal(direct)
	if err != nil {
		t.Fatalf("marshal direct semantic signature: %v", err)
	}
	if string(directData) != want {
		t.Fatalf("unexpected normalized semantic wire:\n got: %s\nwant: %s", directData, want)
	}
}

func TestVectorSearchQueryDefaultsAndOptionalWire(t *testing.T) {
	query, err := NewVectorSearchQuery(VectorSearchQueryInput{Text: "storm warning"})
	if err != nil {
		t.Fatalf("new lexical vector search: %v", err)
	}
	data, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("marshal lexical vector search: %v", err)
	}
	want := `{"text":"storm warning","semantic":null,"minScore":null,"nearbyBucketRadius":1,"maxCandidates":1000,"requireAllTerms":true}`
	if string(data) != want {
		t.Fatalf("unexpected default vector search wire:\n got: %s\nwant: %s", data, want)
	}

	signature, err := NewSemanticVectorSignature(17, 1, []int{1}, []int{2}, []int64{-1})
	if err != nil {
		t.Fatalf("new semantic signature: %v", err)
	}
	minScore := 0.42
	radius := 0
	requireAllTerms := false
	hybrid, err := NewVectorSearchQuery(VectorSearchQueryInput{
		Text:               "hybrid",
		Semantic:           &signature,
		MinScore:           &minScore,
		NearbyBucketRadius: &radius,
		MaxCandidates:      321,
		RequireAllTerms:    &requireAllTerms,
	})
	if err != nil {
		t.Fatalf("new hybrid vector search: %v", err)
	}
	hybridData, err := json.Marshal(hybrid)
	if err != nil {
		t.Fatalf("marshal hybrid vector search: %v", err)
	}
	hybridWant := `{"text":"hybrid","semantic":{"calibrationId":"17","bucketId":1,"cells":[1],"cellCounts":[2],"fingerprint":["0xffffffffffffffff"],"bands":["0x000000000000ffff","0x000000000000ffff","0x000000000000ffff","0x000000000000ffff"],"boundaryConfidence":0},"minScore":0.42,"nearbyBucketRadius":0,"maxCandidates":321,"requireAllTerms":false}`
	if string(hybridData) != hybridWant {
		t.Fatalf("unexpected hybrid vector search wire:\n got: %s\nwant: %s", hybridData, hybridWant)
	}
}

func TestHNSWSearchQueryDefaultsAndLosslessWire(t *testing.T) {
	query, err := NewHNSWSearchQuery(HNSWSearchQueryInput{
		CalibrationID: math.MinInt64,
		Vector:        []float64{0.25, -0.5, 0.75},
		MaxCandidates: 1_200,
	})
	if err != nil {
		t.Fatalf("new HNSW query: %v", err)
	}
	data, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("marshal HNSW query: %v", err)
	}
	want := `{"calibrationId":"-9223372036854775808","vector":[0.25,-0.5,0.75],"maxCandidates":1200,"efSearch":1200,"minScore":null,"formatVersion":1}`
	if string(data) != want {
		t.Fatalf("unexpected HNSW wire:\n got: %s\nwant: %s", data, want)
	}
}

func TestNativeSearchUsesFloat32WireDomain(t *testing.T) {
	overflow := float64(math.MaxFloat32) * 2
	if _, err := NewVectorSearchQuery(VectorSearchQueryInput{
		Text:     "overflow",
		MinScore: &overflow,
	}); err == nil || !strings.Contains(err.Error(), "float32 wire domain") {
		t.Fatalf("expected float32 minScore overflow rejection, got %v", err)
	}

	if _, err := NewHNSWSearchQuery(HNSWSearchQueryInput{
		CalibrationID: 1,
		Vector:        []float64{overflow},
	}); err == nil || !strings.Contains(err.Error(), "float32 wire domain") {
		t.Fatalf("expected float32 vector overflow rejection, got %v", err)
	}

	underflow := float64(math.SmallestNonzeroFloat32) / 4
	if _, err := NewHNSWSearchQuery(HNSWSearchQueryInput{
		CalibrationID: 1,
		Vector:        []float64{underflow},
	}); err == nil || !strings.Contains(err.Error(), "non-zero finite norm") {
		t.Fatalf("expected float32-underflowed zero-norm rejection, got %v", err)
	}

	if _, err := NewHNSWSearchQuery(HNSWSearchQueryInput{
		CalibrationID: 1,
		Vector:        []float64{float64(math.SmallestNonzeroFloat32)},
	}); err != nil {
		t.Fatalf("smallest non-zero float32 vector should remain non-zero: %v", err)
	}

	if _, err := NewHNSWSearchQuery(HNSWSearchQueryInput{
		CalibrationID: 1,
		Vector:        []float64{underflow, 1},
	}); err != nil {
		t.Fatalf("an underflowed component must not invalidate a non-zero vector: %v", err)
	}
}

func TestFloat32RangeChecksPreserveOriginalWireValues(t *testing.T) {
	tinyNegative := -float64(math.SmallestNonzeroFloat32) / 4
	roundsToOne := math.Nextafter(1, math.Inf(1))

	for name, confidence := range map[string]float64{
		"negative rounds to zero": tinyNegative,
		"above one rounds to one": roundsToOne,
	} {
		t.Run("semantic "+name, func(t *testing.T) {
			signature, err := NewSemanticVectorSignature(
				1,
				1,
				[]int{1},
				[]int{2},
				[]int64{1},
				confidence,
			)
			if err != nil {
				t.Fatalf("value valid after float32 narrowing was rejected: %v", err)
			}
			if signature.BoundaryConfidence != confidence {
				t.Fatalf("boundaryConfidence wire value changed: got %v want %v", signature.BoundaryConfidence, confidence)
			}
		})
	}

	for name, score := range map[string]float64{
		"negative rounds to zero": tinyNegative,
		"above one rounds to one": roundsToOne,
	} {
		t.Run("HNSW "+name, func(t *testing.T) {
			query, err := NewHNSWSearchQuery(HNSWSearchQueryInput{
				CalibrationID: 1,
				Vector:        []float64{1},
				MinScore:      &score,
			})
			if err != nil {
				t.Fatalf("value valid after float32 narrowing was rejected: %v", err)
			}
			if query.MinScore == nil || *query.MinScore != score {
				t.Fatalf("minScore wire value changed: got %v want %v", query.MinScore, score)
			}
		})
	}

	outsideUpperBound := float64(math.Nextafter32(1, float32(math.Inf(1))))
	if _, err := NewSemanticVectorSignature(
		1,
		1,
		[]int{1},
		[]int{2},
		[]int64{1},
		outsideUpperBound,
	); err == nil || !strings.Contains(err.Error(), "boundaryConfidence") {
		t.Fatalf("expected out-of-range float32 boundaryConfidence rejection, got %v", err)
	}
	if _, err := NewHNSWSearchQuery(HNSWSearchQueryInput{
		CalibrationID: 1,
		Vector:        []float64{1},
		MinScore:      &outsideUpperBound,
	}); err == nil || !strings.Contains(err.Error(), "minScore") {
		t.Fatalf("expected out-of-range float32 minScore rejection, got %v", err)
	}
}

func TestApproximateIndexCandidateQueryScalarAndList(t *testing.T) {
	scalar, err := NewApproximateIndexCandidateQuery("tenant-1")
	if err != nil {
		t.Fatalf("new scalar candidate query: %v", err)
	}
	assertJSON(t, scalar, `{"values":["tenant-1"],"maxCandidates":1000}`)

	list, err := NewApproximateIndexCandidateQuery([]string{"tenant-1", "tenant-2"}, 17)
	if err != nil {
		t.Fatalf("new list candidate query: %v", err)
	}
	assertJSON(t, list, `{"values":["tenant-1","tenant-2"],"maxCandidates":17}`)
}

func TestNativeSearchValidation(t *testing.T) {
	validSignature, err := NewSemanticVectorSignature(1, 1, []int{1}, []int{2}, []int64{1})
	if err != nil {
		t.Fatalf("valid semantic signature: %v", err)
	}

	semanticCases := []struct {
		name      string
		signature SemanticVectorSignature
		contains  string
	}{
		{
			name: "zero calibration",
			signature: SemanticVectorSignature{
				CalibrationID: "0", BucketID: 1, Cells: []int{1}, CellCounts: []int{2},
				Fingerprint: []string{"1"}, Bands: []string{"1", "0", "0", "0"},
			},
			contains: "calibrationId must be non-zero",
		},
		{
			name:      "mixed radix bucket mismatch",
			signature: withSemanticBucket(validSignature, 0),
			contains:  "bucketId does not match",
		},
		{
			name:      "band mismatch",
			signature: withSemanticBand(validSignature, 0, "2"),
			contains:  "bands do not represent",
		},
		{
			name:      "invalid confidence",
			signature: withSemanticConfidence(validSignature, math.NaN()),
			contains:  "boundaryConfidence",
		},
	}
	for _, test := range semanticCases {
		t.Run("semantic "+test.name, func(t *testing.T) {
			if err := test.signature.Validate(); err == nil || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("expected %q validation error, got %v", test.contains, err)
			}
		})
	}

	text := "text"
	vectorCases := []struct {
		name     string
		query    VectorSearchQuery
		contains string
	}{
		{name: "empty", query: VectorSearchQuery{NearbyBucketRadius: 1, MaxCandidates: 1}, contains: "text and/or"},
		{name: "blank text", query: VectorSearchQuery{Text: stringPointer(" "), NearbyBucketRadius: 1, MaxCandidates: 1}, contains: "non-blank"},
		{name: "non-finite score", query: VectorSearchQuery{Text: &text, MinScore: floatPointer(math.Inf(1)), NearbyBucketRadius: 1, MaxCandidates: 1}, contains: "minScore"},
		{name: "negative radius", query: VectorSearchQuery{Text: &text, NearbyBucketRadius: -1, MaxCandidates: 1}, contains: "nearbyBucketRadius"},
		{name: "candidate bound", query: VectorSearchQuery{Text: &text, NearbyBucketRadius: 1, MaxCandidates: 5_001}, contains: "maxCandidates"},
	}
	for _, test := range vectorCases {
		t.Run("vector "+test.name, func(t *testing.T) {
			if err := test.query.Validate(); err == nil || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("expected %q validation error, got %v", test.contains, err)
			}
		})
	}

	hnswCases := []struct {
		name     string
		query    HNSWSearchQuery
		contains string
	}{
		{name: "format", query: validHNSWQuery(2), contains: "formatVersion"},
		{name: "zero calibration", query: withHNSWCalibration(validHNSWQuery(1), "0"), contains: "non-zero"},
		{name: "empty vector", query: withHNSWVector(validHNSWQuery(1), nil), contains: "dimensions"},
		{name: "non-finite vector", query: withHNSWVector(validHNSWQuery(1), []float64{math.NaN()}), contains: "must be finite"},
		{name: "zero norm", query: withHNSWVector(validHNSWQuery(1), []float64{0, 0}), contains: "non-zero finite norm"},
		{name: "candidate bound", query: withHNSWMaxCandidates(validHNSWQuery(1), 5_001), contains: "maxCandidates"},
		{name: "ef below candidates", query: withHNSWEFSearch(validHNSWQuery(1), 0), contains: "efSearch"},
		{name: "score bound", query: withHNSWMinScore(validHNSWQuery(1), 1.1), contains: "minScore"},
	}
	for _, test := range hnswCases {
		t.Run("HNSW "+test.name, func(t *testing.T) {
			if err := test.query.Validate(); err == nil || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("expected %q validation error, got %v", test.contains, err)
			}
		})
	}

	var nilValue *int
	approximateCases := []struct {
		name     string
		query    ApproximateIndexCandidateQuery
		contains string
	}{
		{name: "empty", query: ApproximateIndexCandidateQuery{MaxCandidates: 1}, contains: "at least one"},
		{name: "null", query: ApproximateIndexCandidateQuery{Values: []any{nilValue}, MaxCandidates: 1}, contains: "cannot be null"},
		{name: "candidate bound", query: ApproximateIndexCandidateQuery{Values: []any{1}, MaxCandidates: 5_001}, contains: "maxCandidates"},
	}
	for _, test := range approximateCases {
		t.Run("approximate "+test.name, func(t *testing.T) {
			if err := test.query.Validate(); err == nil || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("expected %q validation error, got %v", test.contains, err)
			}
		})
	}
}

func TestNativeSearchConditionWireAndRestrictions(t *testing.T) {
	lexical, err := NewVectorSearchQuery(VectorSearchQueryInput{Text: "bounded", MaxCandidates: 128})
	if err != nil {
		t.Fatalf("new lexical query: %v", err)
	}
	assertJSON(
		t,
		ApproximateSearch(lexical),
		`{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"SEARCH_CANDIDATES","value":{"text":"bounded","semantic":null,"minScore":null,"nearbyBucketRadius":1,"maxCandidates":128,"requireAllTerms":true}}}`,
	)

	hnsw, err := NewHNSWSearchQuery(HNSWSearchQueryInput{
		CalibrationID: -7909761245221418085,
		Vector:        []float64{0.25, -0.5, 0.75},
		MaxCandidates: 40,
		EFSearch:      96,
	})
	if err != nil {
		t.Fatalf("new HNSW query: %v", err)
	}
	assertJSON(
		t,
		HNSWCandidates(hnsw),
		`{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"HNSW_CANDIDATES","value":{"calibrationId":"-7909761245221418085","vector":[0.25,-0.5,0.75],"maxCandidates":40,"efSearch":96,"minScore":null,"formatVersion":1}}}`,
	)
	assertJSON(
		t,
		ApproximateCandidates("corpusId", []string{"a", "b"}, 17),
		`{"conditionType":"SingleCondition","criteria":{"field":"corpusId","operator":"CANDIDATES","value":{"values":["a","b"],"maxCandidates":17}}}`,
	)

	semanticSignature := validSemanticSignature(t)
	semantic, err := NewVectorSearchQuery(VectorSearchQueryInput{Semantic: &semanticSignature})
	if err != nil {
		t.Fatalf("new semantic query: %v", err)
	}
	if _, err := json.Marshal(ApproximateSearch(semantic)); err == nil || !strings.Contains(err.Error(), "text-only") {
		t.Fatalf("expected bounded lexical semantic rejection, got %v", err)
	}
	if _, err := json.Marshal(ApproximateCandidates(" ", "value")); err == nil || !strings.Contains(err.Error(), "must not be blank") {
		t.Fatalf("expected blank candidate attribute rejection, got %v", err)
	}
}

func validSemanticSignature(t *testing.T) SemanticVectorSignature {
	t.Helper()
	signature, err := NewSemanticVectorSignature(1, 1, []int{1}, []int{2}, []int64{1})
	if err != nil {
		t.Fatalf("new semantic signature: %v", err)
	}
	return signature
}

func validHNSWQuery(formatVersion int) HNSWSearchQuery {
	return HNSWSearchQuery{
		CalibrationID: "1",
		Vector:        []float64{1},
		MaxCandidates: 10,
		EFSearch:      10,
		FormatVersion: formatVersion,
	}
}

func withSemanticBucket(signature SemanticVectorSignature, bucketID int) SemanticVectorSignature {
	signature.BucketID = bucketID
	return signature
}

func withSemanticBand(signature SemanticVectorSignature, index int, value string) SemanticVectorSignature {
	signature.Bands = append([]string(nil), signature.Bands...)
	signature.Bands[index] = value
	return signature
}

func withSemanticConfidence(signature SemanticVectorSignature, confidence float64) SemanticVectorSignature {
	signature.BoundaryConfidence = confidence
	return signature
}

func withHNSWCalibration(query HNSWSearchQuery, calibrationID string) HNSWSearchQuery {
	query.CalibrationID = calibrationID
	return query
}

func withHNSWVector(query HNSWSearchQuery, vector []float64) HNSWSearchQuery {
	query.Vector = vector
	return query
}

func withHNSWMaxCandidates(query HNSWSearchQuery, maxCandidates int) HNSWSearchQuery {
	query.MaxCandidates = maxCandidates
	return query
}

func withHNSWEFSearch(query HNSWSearchQuery, efSearch int) HNSWSearchQuery {
	query.EFSearch = efSearch
	return query
}

func withHNSWMinScore(query HNSWSearchQuery, minScore float64) HNSWSearchQuery {
	query.MinScore = &minScore
	return query
}

func stringPointer(value string) *string { return &value }

func floatPointer(value float64) *float64 { return &value }

func assertJSON(t *testing.T, value any, want string) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	if string(data) != want {
		t.Fatalf("unexpected JSON:\n got: %s\nwant: %s", data, want)
	}
}
