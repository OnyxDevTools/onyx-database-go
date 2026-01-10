package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/cmd/onyx-gen-go/generator"
	schemaCmds "github.com/OnyxDevTools/onyx-database-go/cmd/onyx-schema-go/commands"
)

func main() {
	exit(runMain(os.Args, os.Stdout, os.Stderr))
}

var exit = os.Exit
var getwd = os.Getwd

func runMain(args []string, stdout, stderr io.Writer) int {
	return dispatch(args[1:], stdout, stderr)
}

func dispatch(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printRootUsage(stdout)
		return 2
	}

	switch args[0] {
	case "-h", "--help", "help":
		printRootUsage(stdout)
		return 0
	case "gen":
		return runGen(args[1:], stdout, stderr)
	case "schema":
		schemaCmds.Stdout = stdout
		schemaCmds.Stderr = stderr
		return schemaCmds.Dispatch(args[1:])
	default:
		fmt.Fprintf(stderr, "unknown subcommand %q\n", args[0])
		printRootUsage(stderr)
		return 2
	}
}

func printRootUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: onyx-go <subcommand> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  gen       Generate Go client/models")
	fmt.Fprintln(w, "  schema    Schema operations (validate/diff/get/publish)")
}

func runGen(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "init" {
		return runGenInit(args[1:], stdout, stderr)
	}

	var usageBuffer bytes.Buffer
	fs := flag.NewFlagSet("onyx-go gen", flag.ContinueOnError)
	fs.SetOutput(&usageBuffer)

	schemaPath := fs.String("schema", "./api/onyx.schema.json", "path to the onyx.schema.json file")
	source := fs.String("source", "file", "schema source: file or api")
	databaseID := fs.String("database-id", "", "database id used when --source=api")
	outPath := fs.String("out", "./gen/onyx", "output directory for generated files")
	packageName := fs.String("package", "", "package name for generated code (default: onyx)")
	tables := fs.String("tables", "", "comma-separated list of tables to generate")
	timestamps := fs.String("timestamps", "time", "timestamp representation: time or string")
	fs.Usage = func() {
		fmt.Fprintf(&usageBuffer, "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
		if err := writeUsageBuffer(&usageBuffer, stdout); err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
		}
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fs.Usage()
			return 0
		}
		fs.Usage()
		if writeErr := writeUsageBuffer(&usageBuffer, stderr); writeErr != nil {
			fmt.Fprintf(stderr, "%v\n", writeErr)
		}
		return 2
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
		fs.Usage()
		if writeErr := writeUsageBuffer(&usageBuffer, stderr); writeErr != nil {
			fmt.Fprintf(stderr, "%v\n", writeErr)
		}
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "Generated code from %s into %s/*.go\n", opts.SchemaPath, opts.OutPath)
	return 0
}

func runGenInit(args []string, stdout, stderr io.Writer) int {
	var usageBuffer bytes.Buffer
	fs := flag.NewFlagSet("onyx-go gen init", flag.ContinueOnError)
	fs.SetOutput(&usageBuffer)

	anchorPath := fs.String("file", "generate.go", "destination for the go:generate anchor file")
	schemaPath := fs.String("schema", "./api/onyx.schema.json", "path to the onyx.schema.json file")
	source := fs.String("source", "file", "schema source: file or api")
	databaseID := fs.String("database-id", "", "database id used when --source=api")
	outPath := fs.String("out", "./gen/onyx", "output directory for generated files")
	packageName := fs.String("package", "onyx", "package name for generated code (default: onyx)")
	tables := fs.String("tables", "", "comma-separated list of tables to generate")
	timestamps := fs.String("timestamps", "time", "timestamp representation: time or string")

	fs.Usage = func() {
		fmt.Fprintf(&usageBuffer, "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
		if err := writeUsageBuffer(&usageBuffer, stdout); err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
		}
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fs.Usage()
			return 0
		}
		fs.Usage()
		if writeErr := writeUsageBuffer(&usageBuffer, stderr); writeErr != nil {
			fmt.Fprintf(stderr, "%v\n", writeErr)
		}
		return 2
	}

	cfg := anchorConfig{
		SchemaPath:      *schemaPath,
		Source:          *source,
		DatabaseID:      *databaseID,
		OutPath:         *outPath,
		PackageName:     *packageName,
		Tables:          parseTables(*tables),
		TimestampFormat: *timestamps,
	}

	content := buildAnchorFile(cfg)

	if err := os.WriteFile(*anchorPath, []byte(content), 0o644); err != nil {
		fmt.Fprintf(stderr, "failed to write anchor file: %v\n", err)
		return 1
	}

	cwd, err := getwd()
	if err != nil {
		fmt.Fprintf(stderr, "getwd: %v\n", err)
		return 1
	}
	printLayout(stdout, cwd, *anchorPath, cfg)
	return 0
}

type anchorConfig struct {
	SchemaPath      string
	Source          string
	DatabaseID      string
	OutPath         string
	PackageName     string
	Tables          []string
	TimestampFormat string
}

func buildAnchorFile(cfg anchorConfig) string {
	var b strings.Builder
	b.WriteString("// Code generated by onyx-go init; DO NOT EDIT.\n")
	b.WriteString("// Generated at: " + time.Now().UTC().Format(time.RFC3339) + "\n\n")
	b.WriteString("package codegen\n\n")

	b.WriteString("//go:generate onyx-go gen")
	if cfg.Source != "" {
		b.WriteString(fmt.Sprintf(" --source %s", cfg.Source))
	}
	if cfg.SchemaPath != "" {
		b.WriteString(fmt.Sprintf(" --schema %s", cfg.SchemaPath))
	}
	if cfg.Source == "api" && cfg.DatabaseID != "" {
		b.WriteString(fmt.Sprintf(" --database-id %s", cfg.DatabaseID))
	}
	if cfg.OutPath != "" {
		b.WriteString(fmt.Sprintf(" --out %s", cfg.OutPath))
	}
	if cfg.PackageName != "" {
		b.WriteString(fmt.Sprintf(" --package %s", cfg.PackageName))
	}
	if cfg.TimestampFormat != "" && cfg.TimestampFormat != "time" {
		b.WriteString(fmt.Sprintf(" --timestamps %s", cfg.TimestampFormat))
	}
	if len(cfg.Tables) > 0 {
		b.WriteString(" --tables ")
		b.WriteString(strings.Join(cfg.Tables, ","))
	}
	b.WriteString("\n")

	return b.String()
}

func printLayout(w io.Writer, cwd, anchorPath string, cfg anchorConfig) {
	fmt.Fprintf(w, "Wrote go:generate anchor to %s\n", anchorPath)
	fmt.Fprintln(w, "You are now ready to generate an onyx client.")
	fmt.Fprintln(w, "The current configuration assumes a folder structure like this:")

	rel := func(p string) string {
		if filepath.IsAbs(p) {
			if r, err := filepath.Rel(cwd, p); err == nil {
				return r
			}
			return p
		}
		return filepath.Clean(p)
	}

	schema := rel(cfg.SchemaPath)
	outDir := rel(cfg.OutPath)
	anchor := rel(anchorPath)

	fmt.Fprintln(w, ".")
	fmt.Fprintf(w, "├── %s\n", anchor)
	fmt.Fprintln(w, "├── api")
	fmt.Fprintf(w, "│   └── %s\n", filepath.Base(schema))
	fmt.Fprintf(w, "└── %s\n", filepath.Dir(outDir))
	fmt.Fprintf(w, "    └── %s\n", filepath.Base(outDir))
	fmt.Fprintln(w, "        ├── common.go")
	fmt.Fprintln(w, "        └── {table}.go")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next Steps:")
	fmt.Fprintln(w, "  1) Generate your typed onyx client:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "     go generate")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  2) Initialize your database connection:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "     import \"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx\"")
	fmt.Fprintln(w, "     db, err := onyx.New(context.Background(), onyx.Config{})")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  3) Start writing code:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "     user, err := db.Users().Save(ctx, onyx.User{})")
}

func writeUsageBuffer(buf *bytes.Buffer, dst io.Writer) error {
	if _, err := buf.WriteTo(dst); err != nil {
		return fmt.Errorf("write usage: %w", err)
	}
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
