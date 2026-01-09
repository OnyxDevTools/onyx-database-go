package schema

import (
	"sort"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

// FieldDiff captures a change to a field.
type FieldDiff struct {
	Name string         `json:"name"`
	From contract.Field `json:"from"`
	To   contract.Field `json:"to"`
}

// TableDiff captures changes within a table.
type TableDiff struct {
	Name             string           `json:"name"`
	AddedFields      []contract.Field `json:"addedFields,omitempty"`
	RemovedFields    []contract.Field `json:"removedFields,omitempty"`
	ModifiedFields   []FieldDiff      `json:"modifiedFields,omitempty"`
	AddedResolvers   []string         `json:"addedResolvers,omitempty"`
	RemovedResolvers []string         `json:"removedResolvers,omitempty"`
	ModifiedResolvers []ResolverDiff  `json:"modifiedResolvers,omitempty"`
}

// SchemaDiff reports differences between schemas.
type SchemaDiff struct {
	AddedTables   []contract.Table `json:"addedTables,omitempty"`
	RemovedTables []contract.Table `json:"removedTables,omitempty"`
	TableDiffs    []TableDiff      `json:"tableDiffs,omitempty"`
}

// ResolverDiff captures a resolver change.
type ResolverDiff struct {
	Name string             `json:"name"`
	From contract.Resolver  `json:"from"`
	To   contract.Resolver  `json:"to"`
}

// DiffSchemas compares two schemas and reports structural differences.
func DiffSchemas(a, b contract.Schema) SchemaDiff {
	normalizedA := contract.NormalizeSchema(a)
	normalizedB := contract.NormalizeSchema(b)

	diff := SchemaDiff{}

	tableMapA := map[string]contract.Table{}
	for _, table := range normalizedA.Tables {
		tableMapA[table.Name] = table
	}

	tableMapB := map[string]contract.Table{}
	for _, table := range normalizedB.Tables {
		tableMapB[table.Name] = table
	}

	for name, tableB := range tableMapB {
		if _, exists := tableMapA[name]; !exists {
			diff.AddedTables = append(diff.AddedTables, tableB)
		}
	}

	for name, tableA := range tableMapA {
		other, exists := tableMapB[name]
		if !exists {
			diff.RemovedTables = append(diff.RemovedTables, tableA)
			continue
		}

		if td := diffTable(tableA, other); td != nil {
			diff.TableDiffs = append(diff.TableDiffs, *td)
		}
	}

	sort.Slice(diff.AddedTables, func(i, j int) bool { return diff.AddedTables[i].Name < diff.AddedTables[j].Name })
	sort.Slice(diff.RemovedTables, func(i, j int) bool { return diff.RemovedTables[i].Name < diff.RemovedTables[j].Name })
	sort.Slice(diff.TableDiffs, func(i, j int) bool { return diff.TableDiffs[i].Name < diff.TableDiffs[j].Name })

	return diff
}

func diffTable(a, b contract.Table) *TableDiff {
	fieldMapA := map[string]contract.Field{}
	for _, f := range a.Fields {
		fieldMapA[f.Name] = f
	}

	fieldMapB := map[string]contract.Field{}
	for _, f := range b.Fields {
		fieldMapB[f.Name] = f
	}

	td := TableDiff{Name: a.Name}

	for name, fieldB := range fieldMapB {
		if _, exists := fieldMapA[name]; !exists {
			td.AddedFields = append(td.AddedFields, fieldB)
		}
	}

	for name, fieldA := range fieldMapA {
		other, exists := fieldMapB[name]
		if !exists {
			td.RemovedFields = append(td.RemovedFields, fieldA)
			continue
		}

		if fieldChanged(fieldA, other) {
			td.ModifiedFields = append(td.ModifiedFields, FieldDiff{Name: name, From: fieldA, To: other})
		}
	}

	resolverMapA := map[string]contract.Resolver{}
	for _, r := range a.Resolvers {
		resolverMapA[r.Name] = r
	}
	resolverMapB := map[string]contract.Resolver{}
	for _, r := range b.Resolvers {
		resolverMapB[r.Name] = r
	}
	for r := range resolverMapA {
		if _, ok := resolverMapB[r]; !ok {
			td.AddedResolvers = append(td.AddedResolvers, r)
			continue
		}
		if resolverChanged(resolverMapA[r], resolverMapB[r]) {
			td.ModifiedResolvers = append(td.ModifiedResolvers, ResolverDiff{
				Name: r,
				From: resolverMapA[r],
				To:   resolverMapB[r],
			})
		}
	}
	for r := range resolverMapB {
		if _, ok := resolverMapA[r]; !ok {
			td.RemovedResolvers = append(td.RemovedResolvers, r)
		}
	}

	sort.Slice(td.AddedFields, func(i, j int) bool { return td.AddedFields[i].Name < td.AddedFields[j].Name })
	sort.Slice(td.RemovedFields, func(i, j int) bool { return td.RemovedFields[i].Name < td.RemovedFields[j].Name })
	sort.Slice(td.ModifiedFields, func(i, j int) bool { return td.ModifiedFields[i].Name < td.ModifiedFields[j].Name })
	sort.Slice(td.ModifiedResolvers, func(i, j int) bool { return td.ModifiedResolvers[i].Name < td.ModifiedResolvers[j].Name })
	sort.Strings(td.AddedResolvers)
	sort.Strings(td.RemovedResolvers)

	if len(td.AddedFields) == 0 &&
		len(td.RemovedFields) == 0 &&
		len(td.ModifiedFields) == 0 &&
		len(td.AddedResolvers) == 0 &&
		len(td.RemovedResolvers) == 0 &&
		len(td.ModifiedResolvers) == 0 {
		return nil
	}

	return &td
}

func fieldChanged(a, b contract.Field) bool {
	return a.Type != b.Type || a.Nullable != b.Nullable
}

func resolverChanged(a, b contract.Resolver) bool {
	if a.Resolver != b.Resolver {
		return true
	}
	if len(a.Meta) != len(b.Meta) {
		return true
	}
	for k, v := range a.Meta {
		if vb, ok := b.Meta[k]; !ok || vb != v {
			return true
		}
	}
	return false
}
