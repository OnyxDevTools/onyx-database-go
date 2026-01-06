//go:build docs

package examples

import (
	"context"
	"fmt"

	models "github.com/OnyxDevTools/onyx-database-go/examples/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// Seed populates a handful of rows used across the examples.
func Seed(ctx context.Context) error {
	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		return err
	}

	users := []any{
		models.User{ID: "user_alice", Email: "alice@example.com", Username: "alice", Region: "us-east-1", IsActive: true, Profile: models.Profile{AvatarURL: "https://example.com/alice.png"}},
		models.User{ID: "user_bob", Email: "bob@example.com", Username: "bobby", Region: "us-west-2", IsActive: true, Profile: models.Profile{AvatarURL: "https://example.com/bob.png"}},
		models.User{ID: "user_cara", Email: "cara@example.com", Username: "cara", Region: "eu-west-1", IsActive: false, Profile: models.Profile{AvatarURL: "https://example.com/cara.png"}},
	}
	if err := db.BatchSave(ctx, "User", users, 50); err != nil {
		return err
	}
	fmt.Println("seeded users:", users)

	roles := []any{
		models.Role{ID: "role_admin", Name: "admin"},
		models.Role{ID: "role_author", Name: "author"},
		models.Role{ID: "role_editor", Name: "editor"},
	}
	if err := db.BatchSave(ctx, "Role", roles, 50); err != nil {
		return err
	}
	fmt.Println("seeded roles:", roles)

	links := []any{
		models.UserRole{ID: "ur_alice_admin", UserID: "user_alice", RoleID: "role_admin"},
		models.UserRole{ID: "ur_bob_author", UserID: "user_bob", RoleID: "role_author"},
		models.UserRole{ID: "ur_cara_editor", UserID: "user_cara", RoleID: "role_editor"},
	}
	if err := db.BatchSave(ctx, "UserRole", links, 50); err != nil {
		return err
	}
	fmt.Println("seeded user roles:", links)

	orders := []any{
		models.Order{ID: "order_200", Region: "us-east-1", Status: "completed", Total: 128.50, CreatedAt: "2024-01-01T12:00:00Z"},
		models.Order{ID: "order_201", Region: "eu-west-1", Status: "pending", Total: 89.10, CreatedAt: "2024-01-02T09:00:00Z"},
		models.Order{ID: "order_202", Region: "us-west-2", Status: "completed", Total: 304.25, CreatedAt: "2024-01-03T16:45:00Z"},
	}

	if err := db.BatchSave(ctx, "Order", orders, 50); err != nil {
		return err
	}

	fmt.Println("seeded orders:", orders)
	return nil
}
