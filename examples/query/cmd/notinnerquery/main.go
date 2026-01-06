package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	users, err := db.From("User").
		Where(contract.NotWithin("id", db.From("UserRole").Select("userId").Where(contract.Eq("roleId", "role-admin")))).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Users without admin role:", users)

	roles, err := db.From("Role").
		Where(contract.NotWithin("id", db.From("RolePermission").Where(contract.Eq("permissionId", "perm-manage-users")))).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Roles missing perm-manage-users:", roles)
}
