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

	users, err := db.ListUsers().
		Select("id").
		Where(onyx.NotWithin("id", coreClient.From(onyxclient.Tables.UserRole).Select("userId").Where(onyx.Eq("roleId", "role-admin")))).
		ListMaps(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Users without admin role:", users)

	roles, err := db.ListRoles().
		Select("id").
		Where(onyx.NotWithin("id", coreClient.From(onyxclient.Tables.RolePermission).Select("roleId").Where(onyx.Eq("permissionId", "perm-manage-users")))).
		ListMaps(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Roles missing perm-manage-users:", roles)
}
