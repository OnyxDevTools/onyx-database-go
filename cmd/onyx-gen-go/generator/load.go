package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

type schemaClient interface {
	Schema(ctx context.Context) (onyx.Schema, error)
}

var initClient = func(ctx context.Context, databaseID string) (schemaClient, error) {
	return onyx.InitWithDatabaseID(ctx, databaseID)
}

func loadSchema(ctx context.Context, opts Options) (onyx.Schema, error) {
	switch opts.Source {
	case "api":
		return loadSchemaFromAPI(ctx, opts)
	case "file":
		fallthrough
	default:
		return loadSchemaFromFile(opts.SchemaPath, opts.Tables)
	}
}

func loadSchemaFromFile(path string, tables []string) (onyx.Schema, error) {
	if path == "" {
		return onyx.Schema{}, fmt.Errorf("schema path is required")
	}

	normalizedPath := filepath.Clean(path)
	data, err := os.ReadFile(normalizedPath)
	if err != nil {
		return onyx.Schema{}, err
	}

	parsed, err := onyx.ParseSchemaJSON(data)
	if err != nil {
		return onyx.Schema{}, err
	}

	return normalizeAndFilter(parsed, tables), nil
}

func loadSchemaFromAPI(ctx context.Context, opts Options) (onyx.Schema, error) {
	client, err := initClient(ctx, opts.DatabaseID)
	if err != nil {
		return onyx.Schema{}, err
	}

	schema, err := client.Schema(ctx)
	if err != nil {
		return onyx.Schema{}, err
	}

	return normalizeAndFilter(schema, opts.Tables), nil
}

func normalizeAndFilter(schema onyx.Schema, allowedTables []string) onyx.Schema {
	if len(allowedTables) == 0 {
		return onyx.NormalizeSchema(schema)
	}

	allowed := map[string]struct{}{}
	for _, name := range allowedTables {
		allowed[name] = struct{}{}
	}

	filtered := onyx.Schema{}
	for _, table := range schema.Tables {
		if _, ok := allowed[table.Name]; ok {
			filtered.Tables = append(filtered.Tables, table)
		}
	}

	return onyx.NormalizeSchema(filtered)
}
