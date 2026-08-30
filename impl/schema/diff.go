package schema

import (
	"github.com/OnyxDevTools/onyx-database-go/contract"
	internal "github.com/OnyxDevTools/onyx-database-go/internal/schema"
)

// SchemaDiff re-exports the schema diff structure.
type SchemaDiff = internal.SchemaDiff

// TableDiff re-exports per-table differences.
type TableDiff = internal.TableDiff

// FieldDiff re-exports field-level differences.
type FieldDiff = internal.FieldDiff

// TableTypeDiff re-exports table management-type differences.
type TableTypeDiff = internal.TableTypeDiff

// SearchSupportDiff re-exports searchable table capability differences.
type SearchSupportDiff = internal.SearchSupportDiff

// DiffSchemas reports the differences between two schemas.
func DiffSchemas(a, b contract.Schema) SchemaDiff {
	return internal.DiffSchemas(a, b)
}
