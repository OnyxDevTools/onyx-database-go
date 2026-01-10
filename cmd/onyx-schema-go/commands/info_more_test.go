package commands

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func TestInfoCommandFlagError(t *testing.T) {
	Stdout, Stderr = &bytes.Buffer{}, &bytes.Buffer{}
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &InfoCommand{}
	if code := cmd.Run([]string{"--bogus"}); code != 2 {
		t.Fatalf("expected usage code, got %d", code)
	}
}

func TestInfoCommandVerifyFailure(t *testing.T) {
	orig := connectInfoClient
	t.Cleanup(func() { connectInfoClient = orig })
	connectInfoClient = func(ctx context.Context, cfg resolver.ResolvedConfig) (schemaClient, error) {
		return schemaClientFunc(func(context.Context) (onyx.Schema, error) {
			return onyx.Schema{}, errors.New("unreachable")
		}), nil
	}

	restoreEnv := setEnvVars(map[string]string{
		"ONYX_DATABASE_ID":         "db",
		"ONYX_DATABASE_BASE_URL":   "http://example.com",
		"ONYX_DATABASE_API_KEY":    "k",
		"ONYX_DATABASE_API_SECRET": "s",
	})
	t.Cleanup(restoreEnv)

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &InfoCommand{}
	if code := cmd.Run(nil); code != 0 {
		t.Fatalf("expected success code with failed status, got %d", code)
	}
	if !bytes.Contains(out.Bytes(), []byte("failed:")) {
		t.Fatalf("expected failed connection status, got %s", out.String())
	}
}

func TestInfoCommandReportsConfigPath(t *testing.T) {
	tmp := t.TempDir()
	configPath := writeConfig(t, tmp, `{"databaseId":"db","databaseBaseUrl":"http://example.com","apiKey":"k","apiSecret":"s"}`)

	restoreEnv := setEnvVars(map[string]string{
		"ONYX_DATABASE_ID":         "",
		"ONYX_DATABASE_BASE_URL":   "",
		"ONYX_DATABASE_API_KEY":    "",
		"ONYX_DATABASE_API_SECRET": "",
	})
	t.Cleanup(restoreEnv)

	orig := connectInfoClient
	t.Cleanup(func() { connectInfoClient = orig })
	connectInfoClient = func(ctx context.Context, cfg resolver.ResolvedConfig) (schemaClient, error) {
		return schemaClientFunc(func(context.Context) (onyx.Schema, error) {
			return onyx.Schema{}, nil
		}), nil
	}

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &InfoCommand{}
	if code := cmd.Run([]string{"--config", configPath}); code != 0 {
		t.Fatalf("expected success, got %d", code)
	}
	if !bytes.Contains(out.Bytes(), []byte(configPath)) {
		t.Fatalf("expected config path in output, got %s", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("ok")) {
		t.Fatalf("expected ok status, got %s", out.String())
	}
}

type schemaClientFunc func(ctx context.Context) (onyx.Schema, error)

func (f schemaClientFunc) Schema(ctx context.Context) (onyx.Schema, error) {
	return f(ctx)
}

func writeConfig(t *testing.T, dir, contents string) string {
	t.Helper()
	path := filepath.Join(dir, "config.json")
	writeFile(t, path, contents)
	return path
}

func setEnvVars(vars map[string]string) func() {
	old := map[string]*string{}
	for k, v := range vars {
		if val, ok := os.LookupEnv(k); ok {
			old[k] = &val
		} else {
			old[k] = nil
		}
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
	return func() {
		for k, v := range old {
			if v == nil {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, *v)
			}
		}
	}
}
