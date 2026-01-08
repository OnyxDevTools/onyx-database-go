package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// PublishCommand pushes a local schema to the API.
type PublishCommand struct{}

func (c *PublishCommand) Name() string        { return "publish" }
func (c *PublishCommand) Description() string { return "publish schema to the API" }

func (c *PublishCommand) Run(args []string) int {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(Stderr)
	databaseID := fs.String("database-id", "", "database id (optional if configured)")
	schemaPath := fs.String("schema", "./onyx.schema.json", "path to schema JSON file")

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

	ctx := context.Background()
	client, err := initSchemaClient(ctx, *databaseID)
	if err != nil {
		fmt.Fprintln(Stderr, err)
		return 1
	}

	if err := client.PublishSchema(ctx, schema); err != nil {
		fmt.Fprintln(Stderr, err)
		return 1
	}

	fmt.Fprintln(Stdout, "Schema published successfully.")
	return 0
}
