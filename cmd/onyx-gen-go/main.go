package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/cmd/onyx-gen-go/generator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	var usageBuffer bytes.Buffer
	fs := flag.NewFlagSet("onyx-gen-go", flag.ContinueOnError)
	fs.SetOutput(&usageBuffer)

	schemaPath := fs.String("schema", "./onyx.schema.json", "path to the onyx.schema.json file")
	source := fs.String("source", "file", "schema source: file or api")
	databaseID := fs.String("database-id", "", "database id used when --source=api")
	outPath := fs.String("out", "./onyxdb", "output directory for generated files (models.go, client.go)")
	packageName := fs.String("package", "", "package name for generated code (defaults to output directory name)")
	tables := fs.String("tables", "", "comma-separated list of tables to generate")
	timestamps := fs.String("timestamps", "time", "timestamp representation: time or string")

	fs.Usage = func() {
		fmt.Fprintf(&usageBuffer, "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
		usageBuffer.WriteTo(stdout)
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fs.Usage()
			return nil
		}

		usageBuffer.WriteTo(stderr)
		return err
	}

	opts := generator.Options{
		SchemaPath:      *schemaPath,
		Source:          *source,
		DatabaseID:      *databaseID,
		OutPath:         *outPath,
		PackageName:     *packageName,
		Tables:          parseTables(*tables),
		TimestampFormat: *timestamps,
	}

	if err := generator.Run(opts); err != nil {
		usageBuffer.WriteTo(stderr)
		return err
	}

	fmt.Fprintf(stdout, "Generated models from %s to %s/models.go\n", opts.SchemaPath, opts.OutPath)
	fmt.Fprintf(stdout, "Generated typed client to %s/client.go\n", opts.OutPath)
	return nil
}

func parseTables(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
