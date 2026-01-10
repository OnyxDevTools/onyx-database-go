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
	exit(runMain(os.Args, os.Stdout, os.Stderr))
}

var exit = os.Exit

func runMain(args []string, stdout, stderr io.Writer) int {
	if err := run(args[1:], stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer) error {
	var usageBuffer bytes.Buffer
	fs := flag.NewFlagSet("onyx-gen-go", flag.ContinueOnError)
	fs.SetOutput(&usageBuffer)

	schemaPath := fs.String("schema", "./api/onyx.schema.json", "path to the onyx.schema.json file")
	source := fs.String("source", "file", "schema source: file or api")
	databaseID := fs.String("database-id", "", "database id used when --source=api")
	outPath := fs.String("out", "./gen/onyx", "output directory for generated files (models.go, client.go)")
	packageName := fs.String("package", "", "package name for generated code (default: onyx)")
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

	fmt.Fprintf(stdout, "Generated code from %s into %s/*.go\n", opts.SchemaPath, opts.OutPath)
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
