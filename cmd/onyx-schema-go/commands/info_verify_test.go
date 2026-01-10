package commands

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
)

func TestInfoCommandResolveErrorAlt(t *testing.T) {
	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	// Point to a missing config path to force resolver error.
	cmd := &InfoCommand{}
	code := cmd.Run([]string{"--config", "missing.json"})
	if code != 1 {
		t.Fatalf("expected resolve error exit code, got %d", code)
	}
}

func TestInfoCommandNoVerify(t *testing.T) {
	orig := connectInfoClient
	t.Cleanup(func() { connectInfoClient = orig })
	connectInfoClient = func(ctx context.Context, cfg resolver.ResolvedConfig) (schemaClient, error) {
		return nil, errors.New("should not be called")
	}

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &InfoCommand{}
	// Provide minimal env so resolve succeeds.
	os.Setenv("ONYX_DATABASE_ID", "db")
	os.Setenv("ONYX_DATABASE_BASE_URL", "http://example.com")
	os.Setenv("ONYX_DATABASE_API_KEY", "k")
	os.Setenv("ONYX_DATABASE_API_SECRET", "s")
	code := cmd.Run([]string{"--no-verify"})
	if code != 0 {
		t.Fatalf("expected success with no-verify, got %d", code)
	}
	if !bytes.Contains(out.Bytes(), []byte("skipped (--no-verify)")) {
		t.Fatalf("expected no-verify status in output")
	}
}
