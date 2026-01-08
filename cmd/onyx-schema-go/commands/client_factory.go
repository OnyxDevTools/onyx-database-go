package commands

import (
	"context"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

var initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
	return onyx.InitWithDatabaseID(ctx, databaseID)
}
