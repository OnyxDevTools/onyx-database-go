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

	users, err := db.Users().
		Select("id").
		Where(onyxdb.NotWithin("id", coreClient.From(onyxdb.Tables.UserRole).Select("userId").Where(onyxdb.Eq("roleId", "role-admin")))).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if users == nil {
		log.Fatalf("warning: expected users response")
	}
	fmt.Println("Users without admin role:", users)

	roles, err := db.Roles().
		Select("id").
		Where(onyxdb.NotWithin("id", coreClient.From(onyxdb.Tables.RolePermission).Select("roleId").Where(onyxdb.Eq("permissionId", "perm-manage-users")))).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if roles == nil {
		log.Fatalf("warning: expected roles response")
	}
	fmt.Println("Roles missing perm-manage-users:", roles)
	log.Println("example: completed")
}
