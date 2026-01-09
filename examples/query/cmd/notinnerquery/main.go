package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := core.Typed()
	coreClient := db.Core()

	users, err := db.ListUsers().
		Select("id").
		Where(onyx.NotWithin("id", coreClient.From(onyx.Tables.UserRole).Select("userId").Where(onyx.Eq("roleId", "role-admin")))).
		ListMaps(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if users == nil {
		log.Fatalf("warning: expected users response")
	}
	fmt.Println("Users without admin role:", users)

	roles, err := db.ListRoles().
		Select("id").
		Where(onyx.NotWithin("id", coreClient.From(onyx.Tables.RolePermission).Select("roleId").Where(onyx.Eq("permissionId", "perm-manage-users")))).
		ListMaps(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if roles == nil {
		log.Fatalf("warning: expected roles response")
	}
	fmt.Println("Roles missing perm-manage-users:", roles)
	log.Println("example: completed")
}
