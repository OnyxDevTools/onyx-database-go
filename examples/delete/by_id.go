//go:build docs

package deleteexamples

import (
	"context"
	"fmt"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// ByID deletes a single row by identifier.
func ByID(ctx context.Context) error {
	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		return err
	}

	targetID := "user_123"
	if err := db.Delete(ctx, "User", targetID); err != nil {
		return err
	}

	fmt.Println("deleted user", targetID)
	return nil
}
