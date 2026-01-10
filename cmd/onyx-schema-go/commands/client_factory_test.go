package commands

import (
	"context"
	"testing"
)

func TestInitSchemaClientDefault(t *testing.T) {
	restore := setEnvVars(map[string]string{
		"ONYX_DATABASE_ID":         "db",
		"ONYX_DATABASE_BASE_URL":   "http://example.com",
		"ONYX_DATABASE_API_KEY":    "k",
		"ONYX_DATABASE_API_SECRET": "s",
	})
	t.Cleanup(restore)

	// We only care that the default path is exercised; an error is acceptable.
	_, _ = initSchemaClient(context.Background(), "db")
}
