package generator

import (
	"context"
	"fmt"
	"strings"
)

// Options captures configuration for the Go code generator.
type Options struct {
	SchemaPath      string
	Source          string
	DatabaseID      string
	OutPath         string
	PackageName     string
	Tables          []string
	TimestampFormat string
}

// Run validates the provided options and performs generation.
// The generation implementation will be added in subsequent tasks.
func Run(opts Options) error {
	if err := ValidateOptions(&opts); err != nil {
		return err
	}

	if _, err := loadSchema(context.Background(), opts); err != nil {
		return err
	}

	// Code generation will be implemented in future tasks.
	return nil
}

// ValidateOptions ensures the provided options meet the CLI requirements.
func ValidateOptions(opts *Options) error {
	if opts == nil {
		return fmt.Errorf("options cannot be nil")
	}

	var missing []string

	if opts.OutPath == "" {
		missing = append(missing, "--out")
	}

	if opts.PackageName == "" {
		missing = append(missing, "--package")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required flag(s): %s", strings.Join(missing, ", "))
	}

	if opts.TimestampFormat == "" {
		opts.TimestampFormat = "time"
	}

	if opts.TimestampFormat != "time" && opts.TimestampFormat != "string" {
		return fmt.Errorf("invalid --timestamps value %q (must be \"time\" or \"string\")", opts.TimestampFormat)
	}

	opts.Source = strings.ToLower(strings.TrimSpace(opts.Source))
	if opts.Source == "" {
		opts.Source = "file"
	}

	if opts.Source != "file" && opts.Source != "api" {
		return fmt.Errorf("invalid --source value %q (must be \"file\" or \"api\")", opts.Source)
	}

	return nil
}
