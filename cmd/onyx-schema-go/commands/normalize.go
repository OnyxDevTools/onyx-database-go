package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// NormalizeCommand writes a normalized schema document.
type NormalizeCommand struct{}

func (c *NormalizeCommand) Name() string        { return "normalize" }
func (c *NormalizeCommand) Description() string { return "normalize a schema file" }

func (c *NormalizeCommand) Run(args []string) int {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(Stderr)
	schemaPath := fs.String("schema", "onyx.schema.json", "path to schema JSON file")
	outPath := fs.String("out", "", "destination file path (defaults to stdout)")

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

	normalized := onyx.NormalizeSchema(schema)
	rendered, err := jsonMarshalIndent(normalized, "", "  ")
	if err != nil {
		fmt.Fprintf(Stderr, "failed to encode schema: %v\n", err)
		return 1
	}

	rendered = append(rendered, '\n')

	if *outPath == "" {
		if _, err := Stdout.Write(rendered); err != nil {
			fmt.Fprintf(Stderr, "failed to write output: %v\n", err)
			return 1
		}
		return 0
	}

	if err := os.WriteFile(*outPath, rendered, 0o644); err != nil {
		fmt.Fprintf(Stderr, "failed to write output file: %v\n", err)
		return 1
	}

	fmt.Fprintf(Stdout, "Wrote normalized schema to %s\n", *outPath)
	return 0
}
