package commands

import (
	"context"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

var initSchemaClient = func(ctx context.Context, databaseID string) (contract.Client, error) {
	return onyx.InitWithDatabaseID(ctx, databaseID)
}
