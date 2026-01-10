package impl

import (
	"context"
	"testing"
)

func TestInitWithDatabaseIDUsesEnv(t *testing.T) {
	t.Setenv("ONYX_DATABASE_ID", "db_env")
	t.Setenv("ONYX_DATABASE_BASE_URL", "https://example.com")
	t.Setenv("ONYX_DATABASE_API_KEY", "k")
	t.Setenv("ONYX_DATABASE_API_SECRET", "s")

	_, err := InitWithDatabaseID(context.Background(), "db_env")
	if err != nil {
		t.Fatalf("InitWithDatabaseID err: %v", err)
	}
}
