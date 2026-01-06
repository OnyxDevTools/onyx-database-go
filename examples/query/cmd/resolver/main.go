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

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Seed a role, permissions, user, profile, and link them together so resolvers have data.
	role := map[string]any{
		"id":          "role_admin_resolver",
		"name":        "Admin",
		"description": "Administrators with full access",
		"isSystem":    false,
	}
	if _, err := db.Save(ctx, "Role", role, nil); err != nil {
		log.Fatal(err)
	}

	permRead := map[string]any{"id": "perm_user_read_resolver", "name": "user.read", "description": "get user(s)"}
	permWrite := map[string]any{"id": "perm_user_write_resolver", "name": "user.write", "description": "Create, update, and delete users"}
	if _, err := db.Save(ctx, "Permission", permRead, nil); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Save(ctx, "Permission", permWrite, nil); err != nil {
		log.Fatal(err)
	}

	rolePerms := []any{
		map[string]any{"id": "rp_read_resolver", "roleId": role["id"], "permissionId": permRead["id"]},
		map[string]any{"id": "rp_write_resolver", "roleId": role["id"], "permissionId": permWrite["id"]},
	}
	for _, rp := range rolePerms {
		if _, err := db.Save(ctx, "RolePermission", rp, nil); err != nil {
			log.Fatal(err)
		}
	}

	userID := "resolver-user-1"
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

	profile := map[string]any{
		"id":        "resolver-profile-1",
		"userId":    userID,
		"firstName": "Example",
		"lastName":  "Admin",
		"bio":       "Seeded admin profile",
		"age":       42,
	}
	if _, err := db.Save(ctx, "UserProfile", profile, nil); err != nil {
		log.Fatal(err)
	}

	userRole := map[string]any{
		"id":     "resolver-user-role-1",
		"userId": userID,
		"roleId": role["id"],
	}
	if _, err := db.Save(ctx, "UserRole", userRole, nil); err != nil {
		log.Fatal(err)
	}

	users, err := db.From("User").
		Where(contract.Eq("id", userID)).
		Resolve("profile", "roles.permissions").
		Limit(5).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(users, "", "  ")
	fmt.Println(string(out))
}
