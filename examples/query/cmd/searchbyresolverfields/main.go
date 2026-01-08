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

	core, err := onyx.Init(ctx, onyx.Config{LogRequests: true})
	if err != nil {
		log.Fatal(err)
	}
	db := onyxclient.NewClient(core)

	// Seed data so resolver-based search has results.
	roleID := "resolver-role-admin"
	role := onyxclient.Role{
		Id:          roleID,
		Name:        "Admin",
		Description: strPtr("Administrators with full access"),
		IsSystem:    false,
	}
	if _, err := db.SaveRole(ctx, role); err != nil {
		log.Fatal(err)
	}

	permRead := onyxclient.Permission{Id: "resolver-perm-read", Name: "user.read", Description: strPtr("get user(s)")}
	permWrite := onyxclient.Permission{Id: "resolver-perm-write", Name: "user.write", Description: strPtr("Create, update, and delete users")}
	if _, err := db.SavePermission(ctx, permRead); err != nil {
		log.Fatal(err)
	}
	if _, err := db.SavePermission(ctx, permWrite); err != nil {
		log.Fatal(err)
	}

	rolePerms := []any{
		onyxclient.RolePermission{Id: "resolver-rp-read", RoleId: roleID, PermissionId: permRead.Id},
		onyxclient.RolePermission{Id: "resolver-rp-write", RoleId: roleID, PermissionId: permWrite.Id},
	}
	for _, rp := range rolePerms {
		if _, err := db.SaveRolePermission(ctx, rp.(onyxclient.RolePermission)); err != nil {
			log.Fatal(err)
		}
	}

	userID := "resolver-admin-user"
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

	userRole := onyxclient.UserRole{
		Id:     "resolver-admin-user-role",
		UserId: userID,
		RoleId: roleID,
	}
	if _, err := db.SaveUserRole(ctx, userRole); err != nil {
		log.Fatal(err)
	}

	admins, err := db.ListUsers().
		Where(onyx.Eq("roles.name", "Admin")).
		Resolve("roles").
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(admins, "", "  ")
	fmt.Println(string(out))
}

func strPtr(s string) *string { return &s }
