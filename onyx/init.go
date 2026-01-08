package onyx

import (
	"context"

	"github.com/OnyxDevTools/onyx-database-go/impl"
)

// Init constructs a client using the provided configuration.
func Init(ctx context.Context, cfg Config) (Client, error) {
	return impl.Init(ctx, cfg)
}

// InitWithDatabaseID initializes a client using only the database ID and environment/file configuration.
func InitWithDatabaseID(ctx context.Context, databaseID string) (Client, error) {
	return impl.InitWithDatabaseID(ctx, databaseID)
}

// ClearConfigCache clears any cached configuration state.
func ClearConfigCache() {
	impl.ClearConfigCache()
}
