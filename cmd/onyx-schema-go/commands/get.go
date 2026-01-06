package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

// GetCommand fetches the schema from the API using resolver/init behavior.
type GetCommand struct{}

func (c *GetCommand) Name() string        { return "get" }
func (c *GetCommand) Description() string { return "retrieve schema from the API" }

func (c *GetCommand) Run(args []string) int {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(Stderr)
	databaseID := fs.String("database-id", "", "database id (optional if configured)")
	outPath := fs.String("out", "", "path to write schema JSON (stdout when empty)")

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

	normalized := contract.NormalizeSchema(schema)
	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		fmt.Fprintln(Stderr, err)
		return 1
	}

	if *outPath == "" {
		if _, err := Stdout.Write(append(data, '\n')); err != nil {
			fmt.Fprintln(Stderr, err)
			return 1
		}
		return 0
	}

	if err := os.WriteFile(*outPath, data, 0o644); err != nil {
		fmt.Fprintf(Stderr, "failed to write schema: %v\n", err)
		return 1
	}

	fmt.Fprintf(Stdout, "Schema written to %s\n", *outPath)
	return 0
}
