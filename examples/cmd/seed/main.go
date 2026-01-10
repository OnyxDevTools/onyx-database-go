package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
)

// Seed populates a handful of rows used across the examples.
func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	core := db.Core()

	users := []any{
		onyx.User{Id: "user_alice", Email: "alice@example.com", Username: "alice", IsActive: true},
		onyx.User{Id: "user_bob", Email: "bob@example.com", Username: "bobby", IsActive: true},
		onyx.User{Id: "user_cara", Email: "cara@example.com", Username: "cara", IsActive: false},
	}
	if len(users) == 0 {
		log.Fatalf("warning: expected seed users")
	}
	if err := core.BatchSave(ctx, onyx.Tables.User, users, 50); err != nil {
		log.Fatal(err)
	}
	fmt.Println("seeded users:", users)

	roles := []any{
		onyx.Role{Id: "role_admin", Name: "admin"},
		onyx.Role{Id: "role_author", Name: "author"},
		onyx.Role{Id: "role_editor", Name: "editor"},
	}
	if len(roles) == 0 {
		log.Fatalf("warning: expected seed roles")
	}
	if err := core.BatchSave(ctx, onyx.Tables.Role, roles, 50); err != nil {
		log.Fatal(err)
	}
	fmt.Println("seeded roles:", roles)

	links := []any{
		onyx.UserRole{Id: "ur_alice_admin", UserId: "user_alice", RoleId: "role_admin"},
		onyx.UserRole{Id: "ur_bob_author", UserId: "user_bob", RoleId: "role_author"},
		onyx.UserRole{Id: "ur_cara_editor", UserId: "user_cara", RoleId: "role_editor"},
	}
	if len(links) == 0 {
		log.Fatalf("warning: expected seed user roles")
	}
	if err := core.BatchSave(ctx, onyx.Tables.UserRole, links, 50); err != nil {
		log.Fatal(err)
	}
	fmt.Println("seeded user roles:", links)
	log.Println("example: completed")
}
