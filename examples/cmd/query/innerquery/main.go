package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	coreClient := db.Core()

	adminUsers, err := db.Users().
		Where(onyx.Within("id", coreClient.From(onyx.Tables.UserRole).Select("userId").Where(onyx.Eq("roleId", "role-admin")))).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if adminUsers == nil {
		log.Fatalf("warning: expected admin users response")
	}
	fmt.Println("Users with admin role:", adminUsers)

	rolesWithPermission, err := db.Roles().
		Where(onyx.Within("id", coreClient.From(onyx.Tables.RolePermission).Select("roleId").Where(onyx.Eq("permissionId", "perm-manage-users")))).
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
