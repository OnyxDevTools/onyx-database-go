package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyxdb"
)

func main() {
	ctx := context.Background()

	db, err := onyxdb.New(ctx, onyxdb.Config{})
	if err != nil {
		log.Fatal(err)
	}

	coreClient := db.Core()

	adminUsers, err := db.Users().
		Where(onyxdb.Within("id", coreClient.From(onyxdb.Tables.UserRole).Select("userId").Where(onyxdb.Eq("roleId", "role-admin")))).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if adminUsers == nil {
		log.Fatalf("warning: expected admin users response")
	}
	fmt.Println("Users with admin role:", adminUsers)

	rolesWithPermission, err := db.Roles().
		Where(onyxdb.Within("id", coreClient.From(onyxdb.Tables.RolePermission).Select("roleId").Where(onyxdb.Eq("permissionId", "perm-manage-users")))).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if rolesWithPermission == nil {
		log.Fatalf("warning: expected roles response")
	}
	fmt.Println("Roles containing permission perm-manage-users:", rolesWithPermission)
	log.Println("example: completed")
}
