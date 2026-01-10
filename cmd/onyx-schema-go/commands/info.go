package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// InfoCommand reports resolved configuration and connectivity.
type InfoCommand struct{}

func (c *InfoCommand) Name() string        { return "info" }
func (c *InfoCommand) Description() string { return "show resolved config and connection status" }

var connectInfoClient = defaultConnectInfoClient

func (c *InfoCommand) Run(args []string) int {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(Stderr)
	databaseID := fs.String("database-id", "", "database id (optional; defaults to env/config)")
	configPath := fs.String("config", "", "path to config file (optional)")
	noVerify := fs.Bool("no-verify", false, "skip live connection check")

	fs.Usage = func() {
		fmt.Fprintf(Stdout, "Usage of %s:\n", c.Name())
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx := context.Background()
	resolved, meta, err := resolver.Resolve(ctx, resolver.Config{
		DatabaseID: *databaseID,
		ConfigPath: *configPath,
	})
	if err != nil {
		fmt.Fprintf(Stderr, "failed to resolve config: %v\n", err)
		return 1
	}

	fmt.Fprintf(Stdout, "Database ID: %s (source: %s)\n", resolved.DatabaseID, meta.Sources.DatabaseID)
	fmt.Fprintf(Stdout, "Base URL   : %s (source: %s)\n", resolved.DatabaseBaseURL, meta.Sources.DatabaseBaseURL)
	fmt.Fprintf(Stdout, "API Key    : %s (source: %s)\n", redact(resolved.APIKey), meta.Sources.APIKey)
	fmt.Fprintf(Stdout, "API Secret : %s (source: %s)\n", redact(resolved.APISecret), meta.Sources.APISecret)
	if meta.FilePath != "" {
		fmt.Fprintf(Stdout, "Config file: %s\n", meta.FilePath)
	}

	status := "skipped (--no-verify)"
	if !*noVerify {
		if err := verifyConnection(ctx, resolved); err != nil {
			status = fmt.Sprintf("failed: %v", err)
		} else {
			status = "ok"
		}
	}

	fmt.Fprintf(Stdout, "Connection : %s\n", status)
	return 0
}

type schemaClient interface {
	Schema(context.Context) (onyx.Schema, error)
}

func verifyConnection(ctx context.Context, cfg resolver.ResolvedConfig) error {
	client, err := connectInfoClient(ctx, cfg)
	if err != nil {
		return err
	}
	_, err = client.Schema(ctx)
	return err
}

func defaultConnectInfoClient(ctx context.Context, cfg resolver.ResolvedConfig) (schemaClient, error) {
	return onyx.Init(ctx, onyx.Config{
		DatabaseID:      cfg.DatabaseID,
		DatabaseBaseURL: cfg.DatabaseBaseURL,
		APIKey:          cfg.APIKey,
		APISecret:       cfg.APISecret,
	})
}

func redact(value string) string {
	if value == "" {
		return "(empty)"
	}
	if len(value) <= 4 {
		return "***"
	}
	return fmt.Sprintf("%s...%s", value[:2], value[len(value)-2:])
}
