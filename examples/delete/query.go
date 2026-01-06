//go:build docs

package deleteexamples

import (
	"context"
	"fmt"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// Query demonstrates deleting a set of rows returned from a query.
func Query(ctx context.Context) error {
	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		return err
	}

	sessions, err := db.From("Session").
		Select("id", "expiresAt").
		Where(contract.Lt("expiresAt", "2024-01-01T00:00:00Z")).
		List(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("stale sessions: %+v\n", sessions)

	for _, session := range sessions {
		id, ok := session["id"].(string)
		if !ok {
			continue
		}
		_ = db.Delete(ctx, "Session", id)
	}

	return nil
}
