package onyx

import (
	"strconv"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

// Re-export contract helpers to keep the public surface stable.
func Asc(field string) Sort     { return contract.Asc(field) }
func Desc(field string) Sort    { return contract.Desc(field) }
func Count(expr string) string  { return unaryQueryFunction("count", expr) }
func Sum(expr string) string    { return unaryQueryFunction("sum", expr) }
func Avg(expr string) string    { return unaryQueryFunction("avg", expr) }
func Min(expr string) string    { return unaryQueryFunction("min", expr) }
func Max(expr string) string    { return unaryQueryFunction("max", expr) }
func Median(expr string) string { return unaryQueryFunction("median", expr) }
func Percentile(expr string, percentile float64) string {
	return queryFunction("percentile", normalizeQueryExpr(expr), formatFloatArg(percentile))
}
func Std(expr string) string      { return unaryQueryFunction("std", expr) }
func Variance(expr string) string { return unaryQueryFunction("variance", expr) }
func Upper(expr string) string    { return unaryQueryFunction("upper", expr) }
func Lower(expr string) string    { return unaryQueryFunction("lower", expr) }
func Format(expr, pattern string) string {
	return queryFunction("format", normalizeQueryExpr(expr), quoteQueryString(pattern))
}
func Substring(expr string, from, length int) string {
	return queryFunction("substring", normalizeQueryExpr(expr), strconv.Itoa(from), strconv.Itoa(length))
}
func Replace(expr, pattern, replacement string) string {
	return queryFunction("replace", normalizeQueryExpr(expr), quoteQueryString(pattern), quoteQueryString(replacement))
}
func Eq(field string, value any) Condition         { return contract.Eq(field, value) }
func Neq(field string, value any) Condition        { return contract.Neq(field, value) }
func In(field string, values []any) Condition      { return contract.In(field, values) }
func NotIn(field string, values []any) Condition   { return contract.NotIn(field, values) }
func Between(field string, from, to any) Condition { return contract.Between(field, from, to) }
func Gt(field string, value any) Condition         { return contract.Gt(field, value) }
func Gte(field string, value any) Condition        { return contract.Gte(field, value) }
func Lt(field string, value any) Condition         { return contract.Lt(field, value) }
func Lte(field string, value any) Condition        { return contract.Lte(field, value) }
func Like(field string, pattern any) Condition     { return contract.Like(field, pattern) }
func Contains(field string, value any) Condition   { return contract.Contains(field, value) }
func StartsWith(field string, value any) Condition { return contract.StartsWith(field, value) }
func Search(queryText string, minScore ...float64) Condition {
	return contract.Search(queryText, minScore...)
}
func VectorSearch(searchQuery VectorSearchQuery) Condition {
	return contract.VectorSearch(searchQuery)
}
func ApproximateSearch(searchQuery VectorSearchQuery) Condition {
	return contract.ApproximateSearch(searchQuery)
}
func HNSWCandidates(searchQuery HNSWSearchQuery) Condition {
	return contract.HNSWCandidates(searchQuery)
}
func ApproximateCandidates(field string, valueOrValues any, maxCandidates ...int) Condition {
	return contract.ApproximateCandidates(field, valueOrValues, maxCandidates...)
}
func NewSemanticVectorSignature(
	calibrationID int64,
	bucketID int,
	cells []int,
	cellCounts []int,
	fingerprint []int64,
	boundaryConfidence ...float64,
) (SemanticVectorSignature, error) {
	return contract.NewSemanticVectorSignature(
		calibrationID,
		bucketID,
		cells,
		cellCounts,
		fingerprint,
		boundaryConfidence...,
	)
}
func NewVectorSearchQuery(input VectorSearchQueryInput) (VectorSearchQuery, error) {
	return contract.NewVectorSearchQuery(input)
}
func NewHNSWSearchQuery(input HNSWSearchQueryInput) (HNSWSearchQuery, error) {
	return contract.NewHNSWSearchQuery(input)
}
func NewApproximateIndexCandidateQuery(
	valueOrValues any,
	maxCandidates ...int,
) (ApproximateIndexCandidateQuery, error) {
	return contract.NewApproximateIndexCandidateQuery(valueOrValues, maxCandidates...)
}
func IsNull(field string) Condition                 { return contract.IsNull(field) }
func NotNull(field string) Condition                { return contract.NotNull(field) }
func Within(field string, query Query) Condition    { return contract.Within(field, query) }
func NotWithin(field string, query Query) Condition { return contract.NotWithin(field, query) }
func Cascade(spec string) CascadeSpec               { return contract.Cascade(spec) }
func NewCascadeBuilder() CascadeBuilder             { return contract.NewCascadeBuilder() }
func NewError(code, message string, meta map[string]any) *Error {
	return contract.NewError(code, message, meta)
}
func NormalizeSchema(s Schema) Schema             { return contract.NormalizeSchema(s) }
func ParseSchemaJSON(data []byte) (Schema, error) { return contract.ParseSchemaJSON(data) }

func unaryQueryFunction(name, expr string) string {
	return queryFunction(name, normalizeQueryExpr(expr))
}

func queryFunction(name string, args ...string) string {
	return name + "(" + strings.Join(args, ",") + ")"
}

func normalizeQueryExpr(expr string) string {
	return strings.TrimSpace(expr)
}

func quoteQueryString(value string) string {
	return strconv.Quote(value)
}

func formatFloatArg(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
