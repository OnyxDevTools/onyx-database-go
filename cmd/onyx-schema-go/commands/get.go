package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// GetCommand fetches the schema from the API using resolver/init behavior.
type GetCommand struct{}

func (c *GetCommand) Name() string        { return "get" }
func (c *GetCommand) Description() string { return "retrieve schema from the API" }

func (c *GetCommand) Run(args []string) int {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(Stderr)
	databaseID := fs.String("database-id", "", "database id (optional; defaults to env/config such as onyx-database.json)")
	outPath := fs.String("out", defaultSchemaPath, "path to write schema JSON (stdout when --print is set)")
	printOnly := fs.Bool("print", false, "print schema to stdout without writing to disk")

	fs.Usage = func() {
		fmt.Fprintf(Stdout, "Usage of %s:\n", c.Name())
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx := context.Background()
	client, err := initSchemaClient(ctx, *databaseID)
	if err != nil {
		fmt.Fprintln(Stderr, err)
		return 1
	}

	schema, err := client.Schema(ctx)
	if err != nil {
		fmt.Fprintln(Stderr, err)
		return 1
	}

	normalized := onyx.NormalizeSchema(schema)
	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		fmt.Fprintln(Stderr, err)
		return 1
	}

	if *printOnly || *outPath == "" {
		if _, err := Stdout.Write(append(data, '\n')); err != nil {
			fmt.Fprintln(Stderr, err)
			return 1
		}
		return 0
	}

	if dir := filepath.Dir(*outPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(Stderr, "failed to create schema directory: %v\n", err)
			return 1
		}
	}

	if err := os.WriteFile(*outPath, data, 0o644); err != nil {
		fmt.Fprintf(Stderr, "failed to write schema: %v\n", err)
		return 1
	}

	fmt.Fprintf(Stdout, "Schema written to %s\n", *outPath)
	return 0
}
