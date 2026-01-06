package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

type schemaClient interface {
	Schema(ctx context.Context) (contract.Schema, error)
}

var initClient = func(ctx context.Context, databaseID string) (schemaClient, error) {
	return onyx.InitWithDatabaseID(ctx, databaseID)
}

func loadSchema(ctx context.Context, opts Options) (contract.Schema, error) {
	switch opts.Source {
	case "api":
		return loadSchemaFromAPI(ctx, opts)
	case "file":
		fallthrough
	default:
		return loadSchemaFromFile(opts.SchemaPath, opts.Tables)
	}
}

func loadSchemaFromFile(path string, tables []string) (contract.Schema, error) {
	if path == "" {
		return contract.Schema{}, fmt.Errorf("schema path is required")
	}

	normalizedPath := filepath.Clean(path)
	data, err := os.ReadFile(normalizedPath)
	if err != nil {
		return contract.Schema{}, err
	}

	parsed, err := contract.ParseSchemaJSON(data)
	if err != nil {
		return contract.Schema{}, err
	}

	return normalizeAndFilter(parsed, tables), nil
}

func loadSchemaFromAPI(ctx context.Context, opts Options) (contract.Schema, error) {
	client, err := initClient(ctx, opts.DatabaseID)
	if err != nil {
		return contract.Schema{}, err
	}

	schema, err := client.Schema(ctx)
	if err != nil {
		return contract.Schema{}, err
	}

	return normalizeAndFilter(schema, opts.Tables), nil
}

func normalizeAndFilter(schema contract.Schema, allowedTables []string) contract.Schema {
	if len(allowedTables) == 0 {
		return contract.NormalizeSchema(schema)
	}

	allowed := map[string]struct{}{}
	for _, name := range allowedTables {
		allowed[name] = struct{}{}
	}

	filtered := contract.Schema{}
	for _, table := range schema.Tables {
		if _, ok := allowed[table.Name]; ok {
			filtered.Tables = append(filtered.Tables, table)
		}
	}

	return contract.NormalizeSchema(filtered)
}
