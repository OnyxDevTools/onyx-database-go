package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

type stubInfoClient struct {
	schemaErr error
}

func (s *stubInfoClient) Schema(ctx context.Context) (onyx.Schema, error) {
	return onyx.Schema{}, s.schemaErr
}

func TestInfoCommandOutputsResolvedConfig(t *testing.T) {
	Stdout, Stderr = &bytes.Buffer{}, &bytes.Buffer{}
	defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "onyx-database.json")
	if err := os.WriteFile(configPath, []byte(`{"databaseId":"file_db","databaseBaseUrl":"https://file.example.com","apiKey":"file_key","apiSecret":"file_secret"}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	connectInfoClient = func(ctx context.Context, cfg resolver.ResolvedConfig) (schemaClient, error) {
		return &stubInfoClient{}, nil
	}
	defer func() { connectInfoClient = defaultConnectInfoClient }()

	cmd := &InfoCommand{}
	if code := cmd.Run([]string{"--config", configPath}); code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, Stderr.(*bytes.Buffer).String())
	}

	out := Stdout.(*bytes.Buffer).String()
	if !bytes.Contains([]byte(out), []byte("file_db")) {
		t.Fatalf("expected database id in output, got: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte(configPath)) {
		t.Fatalf("expected config path in output, got: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("Connection : ok")) {
		t.Fatalf("expected connection ok, got: %s", out)
	}
}
