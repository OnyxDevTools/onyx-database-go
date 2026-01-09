package main

import (
	"context"
	"fmt"
	"log"

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

	coreClient := db.Core()

	adminUsers, err := db.ListUsers().
		Where(onyx.Within("id", coreClient.From(onyxclient.Tables.UserRole).Select("userId").Where(onyx.Eq("roleId", "role-admin")))).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if adminUsers == nil {
		log.Fatalf("warning: expected admin users response")
	}
	fmt.Println("Users with admin role:", adminUsers)

	rolesWithPermission, err := db.ListRoles().
		Where(onyx.Within("id", coreClient.From(onyxclient.Tables.RolePermission).Select("roleId").Where(onyx.Eq("permissionId", "perm-manage-users")))).
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
