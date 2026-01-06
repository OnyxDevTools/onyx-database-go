package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{LogRequests: true})
	if err != nil {
		log.Fatal(err)
	}

	// Seed data so resolver-based search has results.
	roleID := "resolver-role-admin"
	role := map[string]any{
		"id":          roleID,
		"name":        "Admin",
		"description": "Administrators with full access",
		"isSystem":    false,
	}
	if _, err := db.Save(ctx, "Role", role, nil); err != nil {
		log.Fatal(err)
	}

	permRead := map[string]any{"id": "resolver-perm-read", "name": "user.read", "description": "get user(s)"}
	permWrite := map[string]any{"id": "resolver-perm-write", "name": "user.write", "description": "Create, update, and delete users"}
	if _, err := db.Save(ctx, "Permission", permRead, nil); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Save(ctx, "Permission", permWrite, nil); err != nil {
		log.Fatal(err)
	}

	rolePerms := []any{
		map[string]any{"id": "resolver-rp-read", "roleId": roleID, "permissionId": permRead["id"]},
		map[string]any{"id": "resolver-rp-write", "roleId": roleID, "permissionId": permWrite["id"]},
	}
	for _, rp := range rolePerms {
		if _, err := db.Save(ctx, "RolePermission", rp, nil); err != nil {
			log.Fatal(err)
		}
	}

	userID := "resolver-admin-user"
	now := time.Now().UTC().Format(time.RFC3339)
	user := map[string]any{
		"id":        userID,
		"username":  "admin-user-1",
		"email":     "admin@example.com",
		"isActive":  true,
		"createdAt": now,
		"updatedAt": now,
	}
	if _, err := db.Save(ctx, "User", user, nil); err != nil {
		log.Fatal(err)
	}

	userRole := map[string]any{
		"id":     "resolver-admin-user-role",
		"userId": userID,
		"roleId": roleID,
	}
	if _, err := db.Save(ctx, "UserRole", userRole, nil); err != nil {
		log.Fatal(err)
	}

	admins, err := db.From("User").
		Where(contract.Eq("roles.name", "Admin")).
		Resolve("roles").
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(admins, "", "  ")
	fmt.Println(string(out))
}
