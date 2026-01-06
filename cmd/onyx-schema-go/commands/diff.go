package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	schemas "github.com/OnyxDevTools/onyx-database-go/onyx/schema"
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
	jsonOut := fs.Bool("json", false, "emit machine-readable JSON diff")

	fs.Usage = func() {
		fmt.Fprintf(Stdout, "Usage of %s:\n", c.Name())
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *pathB == "" {
		fmt.Fprintln(Stderr, "--b is required")
		fs.Usage()
		return 2
	}

	schemaA, err := loadSchema(*pathA)
	if err != nil {
		fmt.Fprintf(Stderr, "failed to read schema --a: %v\n", err)
		return 1
	}

	schemaB, err := loadSchema(*pathB)
	if err != nil {
		fmt.Fprintf(Stderr, "failed to read schema --b: %v\n", err)
		return 1
	}

	diff := schemas.DiffSchemas(schemaA, schemaB)

	if *jsonOut {
		data, err := json.MarshalIndent(diff, "", "  ")
		if err != nil {
			fmt.Fprintf(Stderr, "failed to render diff: %v\n", err)
			return 1
		}
		fmt.Fprintln(Stdout, string(data))
		return 0
	}

	summary := summarizeDiff(diff)
	if summary == "" {
		fmt.Fprintln(Stdout, "Schemas are identical.")
	} else {
		fmt.Fprintln(Stdout, summary)
	}
	return 0
}

func loadSchema(path string) (contract.Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return contract.Schema{}, err
	}
	return contract.ParseSchemaJSON(data)
}

func summarizeDiff(diff schemas.SchemaDiff) string {
	var sections []string

	if len(diff.AddedTables) > 0 {
		var lines []string
		for _, t := range diff.AddedTables {
			lines = append(lines, fmt.Sprintf("- %s", t.Name))
		}
		sections = append(sections, "Added tables:\n"+strings.Join(lines, "\n"))
	}

	if len(diff.RemovedTables) > 0 {
		var lines []string
		for _, t := range diff.RemovedTables {
			lines = append(lines, fmt.Sprintf("- %s", t.Name))
		}
		sections = append(sections, "Removed tables:\n"+strings.Join(lines, "\n"))
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
		if len(lines) > 0 {
			section := fmt.Sprintf("Changes to table %s:\n%s", td.Name, strings.Join(lines, "\n"))
			sections = append(sections, section)
		}
	}

	return strings.Join(sections, "\n")
}

func describeFieldChange(a, b contract.Field) string {
	var parts []string
	if a.Type != b.Type {
		parts = append(parts, fmt.Sprintf("type %s -> %s", a.Type, b.Type))
	}
	if a.Nullable != b.Nullable {
		parts = append(parts, fmt.Sprintf("nullable %t -> %t", a.Nullable, b.Nullable))
	}
	return strings.Join(parts, "; ")
}
