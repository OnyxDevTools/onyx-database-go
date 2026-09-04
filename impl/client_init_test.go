package impl

import (
	"context"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestInitWithDatabaseIDUsesEnv(t *testing.T) {
	t.Setenv("ONYX_DATABASE_ID", "db_env")
	t.Setenv("ONYX_DATABASE_BASE_URL", "https://example.com")
	t.Setenv("ONYX_DATABASE_API_KEY", "k")
	t.Setenv("ONYX_DATABASE_API_SECRET", "s")

	initialized, err := InitWithDatabaseID(context.Background(), "db_env")
	if err != nil {
		t.Fatalf("InitWithDatabaseID err: %v", err)
	}
	if got := initialized.(*client).wireFormat; got != contract.WireFormatMessagePack {
		t.Fatalf("default wire format = %q, want %q", got, contract.WireFormatMessagePack)
	}
}
