package onyx

import (
	"context"

	"github.com/OnyxDevTools/onyx-database-go/core"
	"github.com/OnyxDevTools/onyx-database-go/impl"
	"github.com/OnyxDevTools/onyx-database-go/onyxclient"
)

type client struct {
	core.Client
}

func (c client) Typed() onyxclient.Client {
	return onyxclient.NewClient(c.Client)
}

// Init constructs a client using the provided configuration.
func Init(ctx context.Context, cfg Config) (Client, error) {
	coreClient, err := impl.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return client{Client: coreClient}, nil
}

// InitWithDatabaseID initializes a client using only the database ID and environment/file configuration.
func InitWithDatabaseID(ctx context.Context, databaseID string) (Client, error) {
	coreClient, err := impl.InitWithDatabaseID(ctx, databaseID)
	if err != nil {
		return nil, err
	}
	return client{Client: coreClient}, nil
}

// ClearConfigCache clears any cached configuration state.
func ClearConfigCache() {
	impl.ClearConfigCache()
}
