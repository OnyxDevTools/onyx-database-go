package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := core.Typed()

	// Seed a role, permissions, user, profile, and link them together so resolvers have data.
	role := onyx.Role{
		Id:          "role_admin_resolver",
		Name:        "Admin",
		Description: strPtr("Administrators with full access"),
		IsSystem:    false,
	}
	_, err = db.SaveRole(ctx, role)
	if err != nil {
		log.Fatal(err)
	}

	permRead := onyx.Permission{Id: "perm_user_read_resolver", Name: "user.read", Description: strPtr("get user(s)")}
	permWrite := onyx.Permission{Id: "perm_user_write_resolver", Name: "user.write", Description: strPtr("Create, update, and delete users")}
	_, err = db.SavePermission(ctx, permRead)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.SavePermission(ctx, permWrite)
	if err != nil {
		log.Fatal(err)
	}

	rolePerms := []any{
		onyx.RolePermission{Id: "rp_read_resolver", RoleId: role.Id, PermissionId: permRead.Id},
		onyx.RolePermission{Id: "rp_write_resolver", RoleId: role.Id, PermissionId: permWrite.Id},
	}
	for _, rp := range rolePerms {
		_, err := db.SaveRolePermission(ctx, rp.(onyx.RolePermission))
		if err != nil {
			log.Fatal(err)
		}
	}

	userID := "resolver-user-1"
	now := time.Now().UTC()
	user := onyx.User{
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

	profile := onyx.UserProfile{
		Id:        "resolver-profile-1",
		UserId:    userID,
		FirstName: "Example",
		LastName:  "Admin",
		Bio:       strPtr("Seeded admin profile"),
		Age:       int64Ptr(42),
	}
	_, err = db.SaveUserProfile(ctx, profile)
	if err != nil {
		log.Fatal(err)
	}

	userRole := onyx.UserRole{
		Id:     "resolver-user-role-1",
		UserId: userID,
		RoleId: role.Id,
	}
	_, err = db.SaveUserRole(ctx, userRole)
	if err != nil {
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
	if len(users) == 0 {
		log.Fatalf("warning: expected resolver user to be returned")
	} else {
		if users[0].Profile == nil {
			log.Fatalf("warning: expected resolver profile")
		}
		if users[0].Roles == nil {
			log.Fatalf("warning: expected resolver roles")
		}
	}

	out, _ := json.MarshalIndent(users, "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}

func strPtr(s string) *string { return &s }
func int64Ptr(v int64) *int64 { return &v }
