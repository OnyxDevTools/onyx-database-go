package commands

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// ValidateCommand checks a schema file for correctness.
type ValidateCommand struct{}

func (c *ValidateCommand) Name() string        { return "validate" }
func (c *ValidateCommand) Description() string { return "validate a schema file" }

func (c *ValidateCommand) Run(args []string) int {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(Stderr)
	schemaPath := fs.String("schema", defaultSchemaPath, "path to schema JSON file")

	fs.Usage = func() {
		fmt.Fprintf(Stdout, "Usage of %s:\n", c.Name())
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	data, err := os.ReadFile(*schemaPath)
	if err != nil {
		fmt.Fprintf(Stderr, "failed to read schema: %v\n", err)
		return 1
	}

	schema, err := onyx.ParseSchemaJSON(data)
	if err != nil {
		fmt.Fprintf(Stderr, "failed to parse schema: %v\n", err)
		return 1
	}

	if errs := validateSchema(schema); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(Stderr, e.Error())
		}
		return 1
	}

	fmt.Fprintln(Stdout, "Schema is valid.")
	return 0
}

func validateSchema(s onyx.Schema) []error {
	var errs []error
	tableNames := map[string]struct{}{}

	for _, table := range s.Tables {
		if table.Name == "" {
			errs = append(errs, errors.New("table name cannot be empty"))
			continue
		}
		if _, exists := tableNames[table.Name]; exists {
			errs = append(errs, fmt.Errorf("duplicate table name: %s", table.Name))
			continue
		}
		tableNames[table.Name] = struct{}{}

		fieldNames := map[string]struct{}{}
		for _, field := range table.Fields {
			if field.Name == "" {
				errs = append(errs, fmt.Errorf("table %s has a field with empty name", table.Name))
				continue
			}
			if _, exists := fieldNames[field.Name]; exists {
				errs = append(errs, fmt.Errorf("table %s has duplicate field name: %s", table.Name, field.Name))
				continue
			}
			fieldNames[field.Name] = struct{}{}
		}
	}

	return errs
}
