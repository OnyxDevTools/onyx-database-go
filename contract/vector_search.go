package contract

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

const (
	// HNSWQueryFormatVersion is the only currently supported native-HNSW wire format.
	HNSWQueryFormatVersion = 1

	// DefaultHNSWCandidates is the default bounded HNSW result budget.
	DefaultHNSWCandidates = 1_000
	// DefaultHNSWEFSearch is the default HNSW distance-evaluation budget.
	DefaultHNSWEFSearch = 1_000
	// MaxHNSWCandidates is the largest HNSW result budget accepted by the API.
	MaxHNSWCandidates = 5_000
	// MaxHNSWEFSearch is the largest HNSW distance-evaluation budget accepted by the API.
	MaxHNSWEFSearch = 20_000
	// MaxHNSWVectorDimension is the largest native-HNSW query vector accepted by the API.
	MaxHNSWVectorDimension = 16_384

	// DefaultApproximateIndexCandidates is the default bounded secondary-index visit budget.
	DefaultApproximateIndexCandidates = 1_000
	// MaxApproximateIndexCandidates is the largest bounded secondary-index visit budget.
	MaxApproximateIndexCandidates = 5_000
	// MaxApproximateIndexRouteValues is the largest bounded secondary-index route list.
	MaxApproximateIndexRouteValues = 5_000

	// DefaultVectorSearchCandidates is the default semantic and hybrid search work budget.
	DefaultVectorSearchCandidates = 1_000
	// MaxVectorSearchCandidates is the hard ceiling for semantic and hybrid search work.
	MaxVectorSearchCandidates = 5_000
)

const semanticBandCount = 4

// SemanticVectorSignature is the lossless wire representation of one semantic
// routing signature. CalibrationID is signed decimal text. Fingerprint and
// Bands are unsigned hexadecimal 64-bit words so JSON decoders cannot round
// their values.
type SemanticVectorSignature struct {
	CalibrationID      string   `json:"calibrationId"`
	BucketID           int      `json:"bucketId"`
	Cells              []int    `json:"cells"`
	CellCounts         []int    `json:"cellCounts"`
	Fingerprint        []string `json:"fingerprint"`
	Bands              []string `json:"bands"`
	BoundaryConfidence float64  `json:"boundaryConfidence"`
}

// NewSemanticVectorSignature validates and constructs a canonical semantic
// signature. Fingerprint words use the same signed, two's-complement int64
// representation as the Kotlin client. Four equal-width bands are derived
// automatically.
func NewSemanticVectorSignature(
	calibrationID int64,
	bucketID int,
	cells []int,
	cellCounts []int,
	fingerprint []int64,
	boundaryConfidence ...float64,
) (SemanticVectorSignature, error) {
	confidence := 0.0
	if len(boundaryConfidence) > 1 {
		return SemanticVectorSignature{}, fmt.Errorf("boundaryConfidence accepts at most one value")
	}
	if len(boundaryConfidence) == 1 {
		confidence = boundaryConfidence[0]
	}

	words := make([]uint64, len(fingerprint))
	for index, word := range fingerprint {
		words[index] = uint64(word)
	}
	bands, err := splitSemanticFingerprint(words)
	if err != nil {
		return SemanticVectorSignature{}, err
	}

	signature := SemanticVectorSignature{
		CalibrationID:      strconv.FormatInt(calibrationID, 10),
		BucketID:           bucketID,
		Cells:              append([]int(nil), cells...),
		CellCounts:         append([]int(nil), cellCounts...),
		Fingerprint:        make([]string, len(words)),
		Bands:              make([]string, len(bands)),
		BoundaryConfidence: confidence,
	}
	for index, word := range words {
		signature.Fingerprint[index] = semanticWireWord(word)
	}
	for index, band := range bands {
		signature.Bands[index] = semanticWireWord(band)
	}
	return signature.normalized()
}

// Validate verifies that the signature satisfies the native semantic-search contract.
func (s SemanticVectorSignature) Validate() error {
	_, err := s.normalized()
	return err
}

// MarshalJSON validates and canonicalizes all lossless integer strings.
func (s SemanticVectorSignature) MarshalJSON() ([]byte, error) {
	normalized, err := s.normalized()
	if err != nil {
		return nil, err
	}
	type wire SemanticVectorSignature
	return json.Marshal(wire(normalized))
}

func (s SemanticVectorSignature) normalized() (SemanticVectorSignature, error) {
	calibrationID, err := signedDecimalInt64(s.CalibrationID, "calibrationId")
	if err != nil {
		return SemanticVectorSignature{}, err
	}
	if calibrationID == "0" {
		return SemanticVectorSignature{}, fmt.Errorf("calibrationId must be non-zero")
	}
	if s.BucketID < 0 {
		return SemanticVectorSignature{}, fmt.Errorf("bucketId must be non-negative")
	}
	if int64(s.BucketID) > int64(math.MaxInt32) {
		return SemanticVectorSignature{}, fmt.Errorf("bucketId exceeds the supported int32 domain")
	}
	if len(s.Cells) == 0 {
		return SemanticVectorSignature{}, fmt.Errorf("at least one product cell is required")
	}
	if len(s.CellCounts) != len(s.Cells) {
		return SemanticVectorSignature{}, fmt.Errorf("cellCounts must contain one cardinality per product cell")
	}

	var packedBucket int64
	bucketSpace := int64(1)
	for axis, cell := range s.Cells {
		count := s.CellCounts[axis]
		if count < 2 {
			return SemanticVectorSignature{}, fmt.Errorf("cellCounts[%d] must be at least 2", axis)
		}
		if cell < 0 || cell >= count {
			return SemanticVectorSignature{}, fmt.Errorf("cells[%d] is outside its cell count", axis)
		}
		if bucketSpace > int64(math.MaxInt32)/int64(count) {
			return SemanticVectorSignature{}, fmt.Errorf("product-cell space exceeds the supported int32 bucket domain")
		}
		bucketSpace *= int64(count)
		packedBucket = packedBucket*int64(count) + int64(cell)
	}
	if int64(s.BucketID) != packedBucket {
		return SemanticVectorSignature{}, fmt.Errorf("bucketId does not match the mixed-radix product cells")
	}

	if len(s.Fingerprint) < 1 || len(s.Fingerprint) > 4 {
		return SemanticVectorSignature{}, fmt.Errorf("fingerprint must contain between 1 and 4 64-bit words")
	}
	fingerprint := make([]uint64, len(s.Fingerprint))
	for index, word := range s.Fingerprint {
		fingerprint[index], err = semanticWord(word, fmt.Sprintf("fingerprint[%d]", index))
		if err != nil {
			return SemanticVectorSignature{}, err
		}
	}
	expectedBands, err := splitSemanticFingerprint(fingerprint)
	if err != nil {
		return SemanticVectorSignature{}, err
	}
	if len(s.Bands) != semanticBandCount {
		return SemanticVectorSignature{}, fmt.Errorf("bands must contain exactly four values")
	}
	bands := make([]uint64, len(s.Bands))
	for index, band := range s.Bands {
		bands[index], err = semanticWord(band, fmt.Sprintf("bands[%d]", index))
		if err != nil {
			return SemanticVectorSignature{}, err
		}
		if bands[index] != expectedBands[index] {
			return SemanticVectorSignature{}, fmt.Errorf("bands do not represent four equal portions of the fingerprint")
		}
	}
	wireBoundaryConfidence := float64(float32(s.BoundaryConfidence))
	if !isFiniteFloat32(s.BoundaryConfidence) || wireBoundaryConfidence < 0 || wireBoundaryConfidence > 1 {
		return SemanticVectorSignature{}, fmt.Errorf("boundaryConfidence must be finite and between zero and one")
	}

	normalized := SemanticVectorSignature{
		CalibrationID:      calibrationID,
		BucketID:           s.BucketID,
		Cells:              append([]int(nil), s.Cells...),
		CellCounts:         append([]int(nil), s.CellCounts...),
		Fingerprint:        make([]string, len(fingerprint)),
		Bands:              make([]string, len(expectedBands)),
		BoundaryConfidence: s.BoundaryConfidence,
	}
	for index, word := range fingerprint {
		normalized.Fingerprint[index] = semanticWireWord(word)
	}
	for index, band := range expectedBands {
		normalized.Bands[index] = semanticWireWord(band)
	}
	return normalized, nil
}

// VectorSearchQuery is the canonical lexical, semantic, or hybrid MATCHES
// value. Use NewVectorSearchQuery to apply defaults to optional parameters.
type VectorSearchQuery struct {
	Text               *string                  `json:"text"`
	Semantic           *SemanticVectorSignature `json:"semantic"`
	MinScore           *float64                 `json:"minScore"`
	NearbyBucketRadius int                      `json:"nearbyBucketRadius"`
	MaxCandidates      int                      `json:"maxCandidates"`
	RequireAllTerms    bool                     `json:"requireAllTerms"`
}

// VectorSearchQueryInput holds optional inputs for NewVectorSearchQuery. An
// empty Text is omitted. A nil NearbyBucketRadius defaults to 1, a zero
// MaxCandidates defaults to 1000, and a nil RequireAllTerms defaults to true.
type VectorSearchQueryInput struct {
	Text               string
	Semantic           *SemanticVectorSignature
	MinScore           *float64
	NearbyBucketRadius *int
	MaxCandidates      int
	RequireAllTerms    *bool
}

// NewVectorSearchQuery validates and applies defaults to a lexical, semantic,
// or hybrid vector-managed search request.
func NewVectorSearchQuery(input VectorSearchQueryInput) (VectorSearchQuery, error) {
	var text *string
	if input.Text != "" {
		value := input.Text
		text = &value
	}
	radius := 1
	if input.NearbyBucketRadius != nil {
		radius = *input.NearbyBucketRadius
	}
	maxCandidates := input.MaxCandidates
	if maxCandidates == 0 {
		maxCandidates = DefaultVectorSearchCandidates
	}
	requireAllTerms := true
	if input.RequireAllTerms != nil {
		requireAllTerms = *input.RequireAllTerms
	}

	query := VectorSearchQuery{
		Text:               text,
		Semantic:           input.Semantic,
		MinScore:           copyFloat64(input.MinScore),
		NearbyBucketRadius: radius,
		MaxCandidates:      maxCandidates,
		RequireAllTerms:    requireAllTerms,
	}
	return query.normalized()
}

// Validate verifies that the query satisfies the native vector-search contract.
func (q VectorSearchQuery) Validate() error {
	_, err := q.normalized()
	return err
}

// MarshalJSON validates and emits the complete camelCase wire object.
func (q VectorSearchQuery) MarshalJSON() ([]byte, error) {
	normalized, err := q.normalized()
	if err != nil {
		return nil, err
	}
	type wire VectorSearchQuery
	return json.Marshal(wire(normalized))
}

func (q VectorSearchQuery) normalized() (VectorSearchQuery, error) {
	var text *string
	if q.Text != nil {
		if strings.TrimSpace(*q.Text) == "" {
			return VectorSearchQuery{}, fmt.Errorf("text must be non-blank when supplied")
		}
		value := *q.Text
		text = &value
	}

	var semantic *SemanticVectorSignature
	if q.Semantic != nil {
		normalized, err := q.Semantic.normalized()
		if err != nil {
			return VectorSearchQuery{}, fmt.Errorf("semantic: %w", err)
		}
		semantic = &normalized
	}
	if text == nil && semantic == nil {
		return VectorSearchQuery{}, fmt.Errorf("VectorSearchQuery must contain text and/or a semantic signature")
	}
	if q.MinScore != nil && !isFiniteFloat32(*q.MinScore) {
		return VectorSearchQuery{}, fmt.Errorf("minScore must be finite in the float32 wire domain")
	}
	if q.NearbyBucketRadius < 0 {
		return VectorSearchQuery{}, fmt.Errorf("nearbyBucketRadius must be non-negative")
	}
	if int64(q.NearbyBucketRadius) > int64(math.MaxInt32) {
		return VectorSearchQuery{}, fmt.Errorf("nearbyBucketRadius exceeds the supported int32 domain")
	}
	if q.MaxCandidates < 1 || q.MaxCandidates > MaxVectorSearchCandidates {
		return VectorSearchQuery{}, fmt.Errorf("maxCandidates must be between 1 and %d", MaxVectorSearchCandidates)
	}

	return VectorSearchQuery{
		Text:               text,
		Semantic:           semantic,
		MinScore:           copyFloat64(q.MinScore),
		NearbyBucketRadius: q.NearbyBucketRadius,
		MaxCandidates:      q.MaxCandidates,
		RequireAllTerms:    q.RequireAllTerms,
	}, nil
}

// HNSWSearchQuery is the canonical bounded native-HNSW candidate request.
// CalibrationID is decimal text to remain lossless through generic JSON decoders.
type HNSWSearchQuery struct {
	CalibrationID string    `json:"calibrationId"`
	Vector        []float64 `json:"vector"`
	MaxCandidates int       `json:"maxCandidates"`
	EFSearch      int       `json:"efSearch"`
	MinScore      *float64  `json:"minScore"`
	FormatVersion int       `json:"formatVersion"`
}

// HNSWSearchQueryInput holds optional inputs for NewHNSWSearchQuery. Zero
// MaxCandidates, EFSearch, and FormatVersion values select their documented defaults.
type HNSWSearchQueryInput struct {
	CalibrationID int64
	Vector        []float64
	MaxCandidates int
	EFSearch      int
	MinScore      *float64
	FormatVersion int
}

// NewHNSWSearchQuery validates and applies defaults to a native-HNSW request.
func NewHNSWSearchQuery(input HNSWSearchQueryInput) (HNSWSearchQuery, error) {
	maxCandidates := input.MaxCandidates
	if maxCandidates == 0 {
		maxCandidates = DefaultHNSWCandidates
	}
	efSearch := input.EFSearch
	if efSearch == 0 {
		efSearch = DefaultHNSWEFSearch
		if maxCandidates > efSearch {
			efSearch = maxCandidates
		}
	}
	formatVersion := input.FormatVersion
	if formatVersion == 0 {
		formatVersion = HNSWQueryFormatVersion
	}
	query := HNSWSearchQuery{
		CalibrationID: strconv.FormatInt(input.CalibrationID, 10),
		Vector:        append([]float64(nil), input.Vector...),
		MaxCandidates: maxCandidates,
		EFSearch:      efSearch,
		MinScore:      copyFloat64(input.MinScore),
		FormatVersion: formatVersion,
	}
	return query.normalized()
}

// Validate verifies that the request satisfies the native-HNSW contract.
func (q HNSWSearchQuery) Validate() error {
	_, err := q.normalized()
	return err
}

// MarshalJSON validates and emits the complete camelCase HNSW wire object.
func (q HNSWSearchQuery) MarshalJSON() ([]byte, error) {
	normalized, err := q.normalized()
	if err != nil {
		return nil, err
	}
	type wire HNSWSearchQuery
	return json.Marshal(wire(normalized))
}

func (q HNSWSearchQuery) normalized() (HNSWSearchQuery, error) {
	if q.FormatVersion != HNSWQueryFormatVersion {
		return HNSWSearchQuery{}, fmt.Errorf(
			"unsupported HNSW query formatVersion %d; expected %d",
			q.FormatVersion,
			HNSWQueryFormatVersion,
		)
	}
	calibrationID, err := signedDecimalInt64(q.CalibrationID, "HNSW calibrationId")
	if err != nil {
		return HNSWSearchQuery{}, err
	}
	if calibrationID == "0" {
		return HNSWSearchQuery{}, fmt.Errorf("HNSW calibrationId must be non-zero")
	}
	if len(q.Vector) < 1 || len(q.Vector) > MaxHNSWVectorDimension {
		return HNSWSearchQuery{}, fmt.Errorf(
			"HNSW vector dimensions must be between 1 and %d",
			MaxHNSWVectorDimension,
		)
	}
	var squaredMagnitude float64
	for index, value := range q.Vector {
		if !isFiniteFloat32(value) {
			return HNSWSearchQuery{}, fmt.Errorf(
				"HNSW vector[%d] must be finite in the float32 wire domain",
				index,
			)
		}
		wireValue := float64(float32(value))
		squaredMagnitude += wireValue * wireValue
	}
	if !isFinite(squaredMagnitude) || squaredMagnitude <= 0 {
		return HNSWSearchQuery{}, fmt.Errorf("HNSW vector must have a non-zero finite norm")
	}
	if q.MaxCandidates < 1 || q.MaxCandidates > MaxHNSWCandidates {
		return HNSWSearchQuery{}, fmt.Errorf(
			"HNSW maxCandidates must be between 1 and %d",
			MaxHNSWCandidates,
		)
	}
	if q.EFSearch < q.MaxCandidates || q.EFSearch > MaxHNSWEFSearch {
		return HNSWSearchQuery{}, fmt.Errorf(
			"HNSW efSearch must be between maxCandidates and %d",
			MaxHNSWEFSearch,
		)
	}
	if q.MinScore != nil {
		wireMinScore := float64(float32(*q.MinScore))
		if !isFiniteFloat32(*q.MinScore) || wireMinScore < -1 || wireMinScore > 1 {
			return HNSWSearchQuery{}, fmt.Errorf("HNSW minScore must be finite and between -1 and 1")
		}
	}

	return HNSWSearchQuery{
		CalibrationID: calibrationID,
		Vector:        append([]float64(nil), q.Vector...),
		MaxCandidates: q.MaxCandidates,
		EFSearch:      q.EFSearch,
		MinScore:      copyFloat64(q.MinScore),
		FormatVersion: q.FormatVersion,
	}, nil
}

// ApproximateIndexCandidateQuery is an explicitly approximate, bounded route
// through one ordinary secondary index.
type ApproximateIndexCandidateQuery struct {
	Values        []any `json:"values"`
	MaxCandidates int   `json:"maxCandidates"`
}

// NewApproximateIndexCandidateQuery creates a bounded EQUAL-style request for
// a scalar or a bounded IN-style request for a slice or array.
func NewApproximateIndexCandidateQuery(
	valueOrValues any,
	maxCandidates ...int,
) (ApproximateIndexCandidateQuery, error) {
	if len(maxCandidates) > 1 {
		return ApproximateIndexCandidateQuery{}, fmt.Errorf("maxCandidates accepts at most one value")
	}
	if existing, ok := valueOrValues.(ApproximateIndexCandidateQuery); ok {
		if len(maxCandidates) == 1 {
			existing.MaxCandidates = maxCandidates[0]
		}
		return existing.normalized()
	}
	if existing, ok := valueOrValues.(*ApproximateIndexCandidateQuery); ok && existing != nil {
		copy := *existing
		if len(maxCandidates) == 1 {
			copy.MaxCandidates = maxCandidates[0]
		}
		return copy.normalized()
	}
	budget := DefaultApproximateIndexCandidates
	if len(maxCandidates) == 1 {
		budget = maxCandidates[0]
	}
	values := candidateValues(valueOrValues)
	query := ApproximateIndexCandidateQuery{Values: values, MaxCandidates: budget}
	return query.normalized()
}

// Validate verifies that the request satisfies the bounded secondary-index contract.
func (q ApproximateIndexCandidateQuery) Validate() error {
	_, err := q.normalized()
	return err
}

// MarshalJSON validates and emits the complete camelCase candidate wire object.
func (q ApproximateIndexCandidateQuery) MarshalJSON() ([]byte, error) {
	normalized, err := q.normalized()
	if err != nil {
		return nil, err
	}
	type wire ApproximateIndexCandidateQuery
	return json.Marshal(wire(normalized))
}

func (q ApproximateIndexCandidateQuery) normalized() (ApproximateIndexCandidateQuery, error) {
	if q.MaxCandidates < 1 || q.MaxCandidates > MaxApproximateIndexCandidates {
		return ApproximateIndexCandidateQuery{}, fmt.Errorf(
			"maxCandidates must be between 1 and %d",
			MaxApproximateIndexCandidates,
		)
	}
	if len(q.Values) == 0 {
		return ApproximateIndexCandidateQuery{}, fmt.Errorf("approximate index candidates require at least one route value")
	}
	if len(q.Values) > MaxApproximateIndexRouteValues {
		return ApproximateIndexCandidateQuery{}, fmt.Errorf(
			"approximate index candidate routes cannot exceed %d values",
			MaxApproximateIndexRouteValues,
		)
	}
	for _, value := range q.Values {
		if isNilJSONValue(value) {
			return ApproximateIndexCandidateQuery{}, fmt.Errorf("approximate index candidate route values cannot be null")
		}
	}
	return ApproximateIndexCandidateQuery{
		Values:        append([]any(nil), q.Values...),
		MaxCandidates: q.MaxCandidates,
	}, nil
}

func signedDecimalInt64(value, field string) (string, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return "", fmt.Errorf("%s must not be blank", field)
	}
	start := 0
	if text[0] == '-' {
		start = 1
	}
	if start == len(text) {
		return "", fmt.Errorf("%s must be a signed decimal 64-bit value", field)
	}
	for _, char := range text[start:] {
		if char < '0' || char > '9' {
			return "", fmt.Errorf("%s must be a signed decimal 64-bit value", field)
		}
	}
	parsed, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return "", fmt.Errorf("%s exceeds the signed 64-bit range", field)
	}
	return strconv.FormatInt(parsed, 10), nil
}

func semanticWord(value, field string) (uint64, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return 0, fmt.Errorf("%s must not be blank", field)
	}
	hexValue := ""
	if strings.HasPrefix(strings.ToLower(text), "0x") {
		hexValue = text[2:]
	} else if strings.IndexFunc(text, func(char rune) bool {
		return (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')
	}) >= 0 {
		hexValue = text
	}
	if hexValue != "" {
		for _, char := range hexValue {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return 0, fmt.Errorf("%s must be signed decimal or unsigned hexadecimal", field)
			}
		}
		parsed, err := strconv.ParseUint(hexValue, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("%s exceeds 64 bits", field)
		}
		return parsed, nil
	}

	normalized, err := signedDecimalInt64(text, field)
	if err != nil {
		return 0, err
	}
	parsed, _ := strconv.ParseInt(normalized, 10, 64)
	return uint64(parsed), nil
}

func semanticWireWord(value uint64) string {
	return fmt.Sprintf("0x%016x", value)
}

func splitSemanticFingerprint(fingerprint []uint64) ([]uint64, error) {
	if len(fingerprint) < 1 || len(fingerprint) > 4 {
		return nil, fmt.Errorf("fingerprint must contain between 1 and 4 64-bit words")
	}
	bandBits := len(fingerprint) * 64 / semanticBandCount
	bands := make([]uint64, semanticBandCount)
	for bandIndex := range bands {
		firstBit := bandIndex * bandBits
		for bandBit := 0; bandBit < bandBits; bandBit++ {
			sourceBit := firstBit + bandBit
			if (fingerprint[sourceBit/64]>>(sourceBit%64))&1 != 0 {
				bands[bandIndex] |= uint64(1) << bandBit
			}
		}
	}
	return bands, nil
}

func candidateValues(valueOrValues any) []any {
	if valueOrValues == nil {
		return []any{nil}
	}
	value := reflect.ValueOf(valueOrValues)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return []any{valueOrValues}
	}
	values := make([]any, value.Len())
	for index := range values {
		values[index] = value.Index(index).Interface()
	}
	return values
}

func isNilJSONValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func copyFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func isFiniteFloat32(value float64) bool {
	return isFinite(value) && !math.IsInf(float64(float32(value)), 0)
}
