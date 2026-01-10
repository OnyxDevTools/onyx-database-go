package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	schemas "github.com/OnyxDevTools/onyx-database-go/impl/schema"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// DiffCommand compares two schema files.
type DiffCommand struct{}

func (c *DiffCommand) Name() string        { return "diff" }
func (c *DiffCommand) Description() string { return "diff two schema files" }

func (c *DiffCommand) Run(args []string) int {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(Stderr)
	pathA := fs.String("a", "onyx.schema.json", "path to base schema JSON")
	pathB := fs.String("b", "", "path to updated schema JSON")
	databaseID := fs.String("database-id", "", "database id to fetch updated schema via API when --b is omitted")
	jsonOut := fs.Bool("json", false, "emit machine-readable JSON diff")

	fs.Usage = func() {
		fmt.Fprintf(Stdout, "Usage of %s:\n", c.Name())
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	var (
		baseSchema    onyx.Schema
		updatedSchema onyx.Schema
		baseSource    string
		updatedSource string
		err           error
	)

	if *pathB != "" {
		baseSource = fmt.Sprintf("base=%s", *pathA)
		baseSchema, err = loadSchema(*pathA)
		if err != nil {
			fmt.Fprintf(Stderr, "failed to read schema --a: %v\n", err)
			return 1
		}

		updatedSource = fmt.Sprintf("updated=%s (file)", *pathB)
		updatedSchema, err = loadSchema(*pathB)
		if err != nil {
			fmt.Fprintf(Stderr, "failed to read schema --b: %v\n", err)
			return 1
		}
	} else {
		updatedSource = fmt.Sprintf("updated=%s (file)", *pathA)
		updatedSchema, err = loadSchema(*pathA)
		if err != nil {
			fmt.Fprintf(Stderr, "failed to read schema --a: %v\n", err)
			return 1
		}

		if *databaseID != "" {
			baseSource = fmt.Sprintf("base=API (database-id=%s)", *databaseID)
		} else {
			baseSource = "base=API (configured credentials)"
		}
		baseSchema, err = fetchSchemaFromAPI(context.Background(), *databaseID)
		if err != nil {
			fmt.Fprintf(Stderr, "failed to fetch schema from API: %v\n", err)
			return 1
		}
	}

	diff := schemas.DiffSchemas(baseSchema, updatedSchema)

	if *jsonOut {
		data, err := jsonMarshalIndent(diff, "", "  ")
		if err != nil {
			fmt.Fprintf(Stderr, "failed to render diff: %v\n", err)
			return 1
		}
		fmt.Fprintln(Stdout, string(data))
		return 0
	}

	fmt.Fprintf(Stdout, "Comparing schemas (%s, %s)\n", baseSource, updatedSource)

	summary := summarizeDiff(diff)
	if summary == "" {
		fmt.Fprintln(Stdout, "Schemas are identical.")
	} else {
		fmt.Fprintln(Stdout, summary)
	}
	return 0
}

func loadSchema(path string) (onyx.Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return onyx.Schema{}, err
	}
	return onyx.ParseSchemaJSON(data)
}

var (
	schemaClientFactoryHandler = func(ctx context.Context, databaseID string) (schemaClient, error) {
		if databaseID != "" {
			return onyx.InitWithDatabaseID(ctx, databaseID)
		}
		return onyx.Init(ctx, onyx.Config{})
	}
	schemaClientFactoryImpl = func(ctx context.Context, databaseID string) (schemaClient, error) {
		return schemaClientFactoryHandler(ctx, databaseID)
	}
	schemaClientFactory = func(ctx context.Context, databaseID string) (schemaClient, error) {
		return schemaClientFactoryImpl(ctx, databaseID)
	}
	connectSchemaClient = func(ctx context.Context, databaseID string) (schemaClient, error) {
		return schemaClientFactory(ctx, databaseID)
	}
	fetchSchemaFromAPI = func(ctx context.Context, databaseID string) (onyx.Schema, error) {
		client, err := connectSchemaClient(ctx, databaseID)
		if err != nil {
			return onyx.Schema{}, err
		}
		return client.Schema(ctx)
	}
)

func summarizeDiff(diff schemas.SchemaDiff) string {
	var sections []string

	if len(diff.AddedTables) > 0 {
		var lines []string
		for _, t := range diff.AddedTables {
			lines = append(lines, fmt.Sprintf("- %s", t.Name))
		}
		sections = append(sections, "Tables only in updated schema:\n"+strings.Join(lines, "\n"))
	}

	if len(diff.RemovedTables) > 0 {
		var lines []string
		for _, t := range diff.RemovedTables {
			lines = append(lines, fmt.Sprintf("- %s", t.Name))
		}
		sections = append(sections, "Tables only in base schema:\n"+strings.Join(lines, "\n"))
	}

	for _, td := range diff.TableDiffs {
		var lines []string
		if len(td.AddedFields) > 0 {
			lines = append(lines, "  Added fields:")
			for _, f := range td.AddedFields {
				lines = append(lines, fmt.Sprintf("  - %s (%s)", f.Name, f.Type))
			}
		}
		if len(td.RemovedFields) > 0 {
			lines = append(lines, "  Removed fields:")
			for _, f := range td.RemovedFields {
				lines = append(lines, fmt.Sprintf("  - %s (%s)", f.Name, f.Type))
			}
		}
		if len(td.ModifiedFields) > 0 {
			lines = append(lines, "  Modified fields:")
			for _, f := range td.ModifiedFields {
				changes := describeFieldChange(f.From, f.To)
				lines = append(lines, fmt.Sprintf("  - %s: %s", f.Name, changes))
			}
		}
		if len(td.ModifiedResolvers) > 0 {
			lines = append(lines, "  Modified resolvers:")
			for _, r := range td.ModifiedResolvers {
				changes := describeResolverChange(r.From, r.To)
				lines = append(lines, fmt.Sprintf("  - %s: %s", r.Name, changes))
			}
		}
		if len(td.AddedResolvers) > 0 {
			lines = append(lines, "  Added resolvers:")
			for _, r := range td.AddedResolvers {
				lines = append(lines, fmt.Sprintf("  - %s", r))
			}
		}
		if len(td.RemovedResolvers) > 0 {
			lines = append(lines, "  Removed resolvers:")
			for _, r := range td.RemovedResolvers {
				lines = append(lines, fmt.Sprintf("  - %s", r))
			}
		}
		if len(lines) > 0 {
			section := fmt.Sprintf("Changes to table %s:\n%s", td.Name, strings.Join(lines, "\n"))
			sections = append(sections, section)
		}
	}

	return strings.Join(sections, "\n")
}

func describeFieldChange(a, b onyx.Field) string {
	var parts []string
	if a.Type != b.Type {
		parts = append(parts, fmt.Sprintf("type %s -> %s", a.Type, b.Type))
	}
	if a.Nullable != b.Nullable {
		parts = append(parts, fmt.Sprintf("nullable %t -> %t", a.Nullable, b.Nullable))
	}
	return strings.Join(parts, "; ")
}

func describeResolverChange(a, b onyx.Resolver) string {
	var parts []string
	if a.Resolver != b.Resolver {
		parts = append(parts, "definition changed")
	}
	if len(a.Meta) != len(b.Meta) {
		parts = append(parts, "meta changed")
	} else {
		for k, v := range a.Meta {
			if vb, ok := b.Meta[k]; !ok || vb != v {
				parts = append(parts, "meta changed")
				break
			}
		}
	}
	if len(parts) == 0 {
		return "unchanged"
	}
	return strings.Join(parts, "; ")
}
