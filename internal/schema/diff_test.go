package schema

import (
	"reflect"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestDiffSchemasDetectsChanges(t *testing.T) {
	base := contract.Schema{
		Tables: []contract.Table{
			{
				Name: "users",
				Fields: []contract.Field{
					{Name: "id", Type: "string"},
					{Name: "name", Type: "string"},
				},
				Resolvers: []contract.Resolver{
					{Name: "by_email", Resolver: "sql", Meta: map[string]any{"env": "prod"}},
					{Name: "legacy", Resolver: "legacy"},
				},
			},
			{
				Name:   "removed",
				Fields: []contract.Field{{Name: "id", Type: "string"}},
			},
		},
	}

	updated := contract.Schema{
		Tables: []contract.Table{
			{
				Name: "users",
				Fields: []contract.Field{
					{Name: "id", Type: "string", Nullable: true},
					{Name: "email", Type: "string"},
				},
				Resolvers: []contract.Resolver{
					{Name: "by_email", Resolver: "sql_v2", Meta: map[string]any{"env": "prod", "version": "2"}},
					{Name: "new", Resolver: "latest"},
				},
			},
			{
				Name:   "added",
				Fields: []contract.Field{{Name: "id", Type: "string"}},
			},
		},
	}

	diff := DiffSchemas(base, updated)
	if len(diff.AddedTables) != 1 || diff.AddedTables[0].Name != "added" {
		t.Fatalf("expected added table 'added', got %+v", diff.AddedTables)
	}
	if len(diff.RemovedTables) != 1 || diff.RemovedTables[0].Name != "removed" {
		t.Fatalf("expected removed table 'removed', got %+v", diff.RemovedTables)
	}
	if len(diff.TableDiffs) != 1 {
		t.Fatalf("expected table diff for users, got %+v", diff.TableDiffs)
	}

	users := diff.TableDiffs[0]
	if users.Name != "users" {
		t.Fatalf("expected users diff, got %s", users.Name)
	}
	if len(users.AddedFields) != 1 || users.AddedFields[0].Name != "email" {
		t.Fatalf("expected added email field, got %+v", users.AddedFields)
	}
	if len(users.RemovedFields) != 1 || users.RemovedFields[0].Name != "name" {
		t.Fatalf("expected removed name field, got %+v", users.RemovedFields)
	}
	if len(users.ModifiedFields) != 1 || users.ModifiedFields[0].Name != "id" {
		t.Fatalf("expected modified id field, got %+v", users.ModifiedFields)
	}
	if len(users.AddedResolvers) != 1 || users.AddedResolvers[0] != "legacy" {
		t.Fatalf("expected added resolver, got %+v", users.AddedResolvers)
	}
	if len(users.RemovedResolvers) != 1 || users.RemovedResolvers[0] != "new" {
		t.Fatalf("expected removed resolver, got %+v", users.RemovedResolvers)
	}
	if len(users.ModifiedResolvers) != 1 || users.ModifiedResolvers[0].Name != "by_email" {
		t.Fatalf("expected modified resolver, got %+v", users.ModifiedResolvers)
	}
}

func TestDiffSchemasNoChanges(t *testing.T) {
	base := contract.Schema{
		Tables: []contract.Table{
			{
				Name: "alpha",
				Fields: []contract.Field{
					{Name: "b", Type: "string"},
					{Name: "a", Type: "string"},
				},
				Resolvers: []contract.Resolver{
					{Name: "res", Resolver: "sql"},
				},
			},
		},
	}
	updated := contract.Schema{
		Tables: []contract.Table{
			{
				Name: "alpha",
				Fields: []contract.Field{
					{Name: "a", Type: "string"},
					{Name: "b", Type: "string"},
				},
				Resolvers: []contract.Resolver{
					{Name: "res", Resolver: "sql"},
				},
			},
		},
	}

	diff := DiffSchemas(base, updated)
	if !reflect.DeepEqual(diff, SchemaDiff{}) {
		t.Fatalf("expected no diff, got %+v", diff)
	}
}

func TestHelperChangeDetectors(t *testing.T) {
	fieldCases := []struct {
		name string
		a    contract.Field
		b    contract.Field
		want bool
	}{
		{name: "same", a: contract.Field{Name: "id", Type: "string"}, b: contract.Field{Name: "id", Type: "string"}, want: false},
		{name: "type", a: contract.Field{Name: "id", Type: "string"}, b: contract.Field{Name: "id", Type: "int"}, want: true},
		{name: "nullable", a: contract.Field{Name: "id", Type: "string"}, b: contract.Field{Name: "id", Type: "string", Nullable: true}, want: true},
	}

	for _, tt := range fieldCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := fieldChanged(tt.a, tt.b); got != tt.want {
				t.Fatalf("expected %t, got %t", tt.want, got)
			}
		})
	}

	resolverCases := []struct {
		name string
		a    contract.Resolver
		b    contract.Resolver
		want bool
	}{
		{name: "same", a: contract.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v"}}, b: contract.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v"}}, want: false},
		{name: "resolver", a: contract.Resolver{Name: "r", Resolver: "sql"}, b: contract.Resolver{Name: "r", Resolver: "sql_v2"}, want: true},
		{name: "meta length", a: contract.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v"}}, b: contract.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v", "k2": "v2"}}, want: true},
		{name: "meta value", a: contract.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v"}}, b: contract.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "other"}}, want: true},
	}

	for _, tt := range resolverCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolverChanged(tt.a, tt.b); got != tt.want {
				t.Fatalf("expected %t, got %t", tt.want, got)
			}
		})
	}
}
