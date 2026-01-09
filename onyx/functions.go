package onyx

import "github.com/OnyxDevTools/onyx-database-go/core"

// Re-export contract helpers to keep the public surface stable.
func Asc(field string) Sort                         { return core.Asc(field) }
func Desc(field string) Sort                        { return core.Desc(field) }
func Eq(field string, value any) Condition          { return core.Eq(field, value) }
func Neq(field string, value any) Condition         { return core.Neq(field, value) }
func In(field string, values []any) Condition       { return core.In(field, values) }
func NotIn(field string, values []any) Condition    { return core.NotIn(field, values) }
func Between(field string, from, to any) Condition  { return core.Between(field, from, to) }
func Gt(field string, value any) Condition          { return core.Gt(field, value) }
func Gte(field string, value any) Condition         { return core.Gte(field, value) }
func Lt(field string, value any) Condition          { return core.Lt(field, value) }
func Lte(field string, value any) Condition         { return core.Lte(field, value) }
func Like(field string, pattern any) Condition      { return core.Like(field, pattern) }
func Contains(field string, value any) Condition    { return core.Contains(field, value) }
func StartsWith(field string, value any) Condition  { return core.StartsWith(field, value) }
func IsNull(field string) Condition                 { return core.IsNull(field) }
func NotNull(field string) Condition                { return core.NotNull(field) }
func Within(field string, query Query) Condition    { return core.Within(field, query) }
func NotWithin(field string, query Query) Condition { return core.NotWithin(field, query) }
func Cascade(spec string) CascadeSpec               { return core.Cascade(spec) }
func NewCascadeBuilder() CascadeBuilder             { return core.NewCascadeBuilder() }
func NewError(code, message string, meta map[string]any) *Error {
	return core.NewError(code, message, meta)
}
func NormalizeSchema(s Schema) Schema             { return core.NormalizeSchema(s) }
func ParseSchemaJSON(data []byte) (Schema, error) { return core.ParseSchemaJSON(data) }
