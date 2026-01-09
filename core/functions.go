package core

import "github.com/OnyxDevTools/onyx-database-go/contract"

// Re-export contract helpers to keep the public surface stable.
func Asc(field string) Sort                         { return contract.Asc(field) }
func Desc(field string) Sort                        { return contract.Desc(field) }
func Eq(field string, value any) Condition          { return contract.Eq(field, value) }
func Neq(field string, value any) Condition         { return contract.Neq(field, value) }
func In(field string, values []any) Condition       { return contract.In(field, values) }
func NotIn(field string, values []any) Condition    { return contract.NotIn(field, values) }
func Between(field string, from, to any) Condition  { return contract.Between(field, from, to) }
func Gt(field string, value any) Condition          { return contract.Gt(field, value) }
func Gte(field string, value any) Condition         { return contract.Gte(field, value) }
func Lt(field string, value any) Condition          { return contract.Lt(field, value) }
func Lte(field string, value any) Condition         { return contract.Lte(field, value) }
func Like(field string, pattern any) Condition      { return contract.Like(field, pattern) }
func Contains(field string, value any) Condition    { return contract.Contains(field, value) }
func StartsWith(field string, value any) Condition  { return contract.StartsWith(field, value) }
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
