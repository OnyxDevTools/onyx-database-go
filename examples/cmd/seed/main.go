package main

import (
	"context"
	"fmt"
	"log"

	model "github.com/OnyxDevTools/onyx-database-go/examples/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// Seed populates a handful of rows used across the examples.
func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	users := []any{
		model.User{Id: "user_alice", Email: "alice@example.com", Username: "alice", IsActive: true},
		model.User{Id: "user_bob", Email: "bob@example.com", Username: "bobby", IsActive: true},
		model.User{Id: "user_cara", Email: "cara@example.com", Username: "cara", IsActive: false},
	}
	if len(users) == 0 {
		log.Println("warning: expected seed users")
	}
	if err := db.BatchSave(ctx, model.Tables.User, users, 50); err != nil {
		log.Fatal(err)
	}
	fmt.Println("seeded users:", users)

	roles := []any{
		model.Role{Id: "role_admin", Name: "admin"},
		model.Role{Id: "role_author", Name: "author"},
		model.Role{Id: "role_editor", Name: "editor"},
	}
	if len(roles) == 0 {
		log.Println("warning: expected seed roles")
	}
	if err := db.BatchSave(ctx, model.Tables.Role, roles, 50); err != nil {
		log.Fatal(err)
	}
	fmt.Println("seeded roles:", roles)

	links := []any{
		model.UserRole{Id: "ur_alice_admin", UserId: "user_alice", RoleId: "role_admin"},
		model.UserRole{Id: "ur_bob_author", UserId: "user_bob", RoleId: "role_author"},
		model.UserRole{Id: "ur_cara_editor", UserId: "user_cara", RoleId: "role_editor"},
	}
	if len(links) == 0 {
		log.Println("warning: expected seed user roles")
	}
	if err := db.BatchSave(ctx, model.Tables.UserRole, links, 50); err != nil {
		log.Fatal(err)
	}
	fmt.Println("seeded user roles:", links)
	log.Println("example: completed")
}
