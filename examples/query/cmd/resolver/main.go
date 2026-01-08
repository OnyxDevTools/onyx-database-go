package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyxclient"
)

func main() {
	ctx := context.Background()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := onyxclient.NewClient(core)

	// Seed a role, permissions, user, profile, and link them together so resolvers have data.
	role := onyxclient.Role{
		Id:          "role_admin_resolver",
		Name:        "Admin",
		Description: strPtr("Administrators with full access"),
		IsSystem:    false,
	}
	if _, err := db.SaveRole(ctx, role); err != nil {
		log.Fatal(err)
	}

	permRead := onyxclient.Permission{Id: "perm_user_read_resolver", Name: "user.read", Description: strPtr("get user(s)")}
	permWrite := onyxclient.Permission{Id: "perm_user_write_resolver", Name: "user.write", Description: strPtr("Create, update, and delete users")}
	if _, err := db.SavePermission(ctx, permRead); err != nil {
		log.Fatal(err)
	}
	if _, err := db.SavePermission(ctx, permWrite); err != nil {
		log.Fatal(err)
	}

	rolePerms := []any{
		onyxclient.RolePermission{Id: "rp_read_resolver", RoleId: role.Id, PermissionId: permRead.Id},
		onyxclient.RolePermission{Id: "rp_write_resolver", RoleId: role.Id, PermissionId: permWrite.Id},
	}
	for _, rp := range rolePerms {
		if _, err := db.SaveRolePermission(ctx, rp.(onyxclient.RolePermission)); err != nil {
			log.Fatal(err)
		}
	}

	userID := "resolver-user-1"
	now := time.Now().UTC()
	user := onyxclient.User{
		Id:        userID,
		Username:  "admin-user-1",
		Email:     "admin@example.com",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := db.SaveUser(ctx, user); err != nil {
		log.Fatal(err)
	}

	profile := onyxclient.UserProfile{
		Id:        "resolver-profile-1",
		UserId:    userID,
		FirstName: "Example",
		LastName:  "Admin",
		Bio:       strPtr("Seeded admin profile"),
		Age:       int64Ptr(42),
	}
	if _, err := db.SaveUserProfile(ctx, profile); err != nil {
		log.Fatal(err)
	}

	userRole := onyxclient.UserRole{
		Id:     "resolver-user-role-1",
		UserId: userID,
		RoleId: role.Id,
	}
	if _, err := db.SaveUserRole(ctx, userRole); err != nil {
		log.Fatal(err)
	}

	users, err := db.ListUsers().
		Where(onyx.Eq("id", userID)).
		Resolve("profile", "roles.permissions").
		Limit(5).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(users, "", "  ")
	fmt.Println(string(out))
}

func strPtr(s string) *string { return &s }
func int64Ptr(v int64) *int64 { return &v }
