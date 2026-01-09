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

	db, err := onyxdb.New(ctx, onyxdb.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Seed a role, permissions, user, profile, and link them together so resolvers have data.
	role := onyxdb.Role{
		Id:          "role_admin_resolver",
		Name:        "Admin",
		Description: strPtr("Administrators with full access"),
		IsSystem:    false,
	}
	_, err = db.Roles().Save(ctx, role)
	if err != nil {
		log.Fatal(err)
	}

	permRead := onyxdb.Permission{Id: "perm_user_read_resolver", Name: "user.read", Description: strPtr("get user(s)")}
	permWrite := onyxdb.Permission{Id: "perm_user_write_resolver", Name: "user.write", Description: strPtr("Create, update, and delete users")}
	_, err = db.Permissions().Save(ctx, permRead)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Permissions().Save(ctx, permWrite)
	if err != nil {
		log.Fatal(err)
	}

	rolePerms := []any{
		onyxdb.RolePermission{Id: "rp_read_resolver", RoleId: role.Id, PermissionId: permRead.Id},
		onyxdb.RolePermission{Id: "rp_write_resolver", RoleId: role.Id, PermissionId: permWrite.Id},
	}
	for _, rp := range rolePerms {
		_, err := db.RolePermissions().Save(ctx, rp.(onyxdb.RolePermission))
		if err != nil {
			log.Fatal(err)
		}
	}

	userID := "resolver-user-1"
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

	profile := onyxdb.UserProfile{
		Id:        "resolver-profile-1",
		UserId:    userID,
		FirstName: "Example",
		LastName:  "Admin",
		Bio:       strPtr("Seeded admin profile"),
		Age:       int64Ptr(42),
	}
	_, err = db.UserProfiles().Save(ctx, profile)
	if err != nil {
		log.Fatal(err)
	}

	userRole := onyxdb.UserRole{
		Id:     "resolver-user-role-1",
		UserId: userID,
		RoleId: role.Id,
	}
	_, err = db.UserRoles().Save(ctx, userRole)
	if err != nil {
		log.Fatal(err)
	}

	users, err := db.Users().
		Where(onyxdb.Eq("id", userID)).
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
