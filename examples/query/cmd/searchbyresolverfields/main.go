package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyxdb"
)

func main() {
	ctx := context.Background()

	db, err := onyxdb.New(ctx, onyxdb.Config{LogRequests: true})
	if err != nil {
		log.Fatal(err)
	}

	// Seed data so resolver-based search has results.
	roleID := "resolver-role-admin"
	role := onyxdb.Role{
		Id:          roleID,
		Name:        "Admin",
		Description: strPtr("Administrators with full access"),
		IsSystem:    false,
	}
	_, err = db.Roles().Save(ctx, role)
	if err != nil {
		log.Fatal(err)
	}

	permRead := onyxdb.Permission{Id: "resolver-perm-read", Name: "user.read", Description: strPtr("get user(s)")}
	permWrite := onyxdb.Permission{Id: "resolver-perm-write", Name: "user.write", Description: strPtr("Create, update, and delete users")}
	_, err = db.Permissions().Save(ctx, permRead)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Permissions().Save(ctx, permWrite)
	if err != nil {
		log.Fatal(err)
	}

	rolePerms := []any{
		onyxdb.RolePermission{Id: "resolver-rp-read", RoleId: roleID, PermissionId: permRead.Id},
		onyxdb.RolePermission{Id: "resolver-rp-write", RoleId: roleID, PermissionId: permWrite.Id},
	}
	for _, rp := range rolePerms {
		_, err := db.RolePermissions().Save(ctx, rp.(onyxdb.RolePermission))
		if err != nil {
			log.Fatal(err)
		}
	}

	userID := "resolver-admin-user"
	now := time.Now().UTC()
	user := onyxdb.User{
		Id:        userID,
		Username:  "admin-user-1",
		Email:     "admin@example.com",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = db.Users().Save(ctx, user)
	if err != nil {
		log.Fatal(err)
	}

	userRole := onyxdb.UserRole{
		Id:     "resolver-admin-user-role",
		UserId: userID,
		RoleId: roleID,
	}
	_, err = db.UserRoles().Save(ctx, userRole)
	if err != nil {
		log.Fatal(err)
	}

	admins, err := db.Users().
		Where(onyxdb.Eq("roles.name", "Admin")).
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
