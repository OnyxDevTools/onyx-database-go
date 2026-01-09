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
	_, err = db.SaveRole(ctx, role)
	if err != nil {
		log.Fatal(err)
	}

	permRead := onyxclient.Permission{Id: "resolver-perm-read", Name: "user.read", Description: strPtr("get user(s)")}
	permWrite := onyxclient.Permission{Id: "resolver-perm-write", Name: "user.write", Description: strPtr("Create, update, and delete users")}
	_, err = db.SavePermission(ctx, permRead)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.SavePermission(ctx, permWrite)
	if err != nil {
		log.Fatal(err)
	}

	rolePerms := []any{
		onyxclient.RolePermission{Id: "resolver-rp-read", RoleId: roleID, PermissionId: permRead.Id},
		onyxclient.RolePermission{Id: "resolver-rp-write", RoleId: roleID, PermissionId: permWrite.Id},
	}
	for _, rp := range rolePerms {
		_, err := db.SaveRolePermission(ctx, rp.(onyxclient.RolePermission))
		if err != nil {
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
	_, err = db.SaveUser(ctx, user)
	if err != nil {
		log.Fatal(err)
	}

	userRole := onyxclient.UserRole{
		Id:     "resolver-admin-user-role",
		UserId: userID,
		RoleId: roleID,
	}
	_, err = db.SaveUserRole(ctx, userRole)
	if err != nil {
		log.Fatal(err)
	}

	admins, err := db.ListUsers().
		Where(onyx.Eq("roles.name", "Admin")).
		Resolve("roles").
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if len(admins) == 0 {
		log.Printf("warning: expected admin users from resolver query")
	} else if admins[0].Roles == nil {
		log.Printf("warning: expected resolved roles")
	}

	out, _ := json.MarshalIndent(admins, "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}

func strPtr(s string) *string { return &s }
