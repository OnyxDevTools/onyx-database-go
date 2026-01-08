package main

import (
	"context"
	"fmt"
	"log"
	"time"

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

	// Seed a record to ensure the delete has a target.
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	seed := onyxclient.User{
		Id:        "obsolete_user_1",
		Username:  "obsolete",
		Email:     "obsolete@example.com",
		IsActive:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = db.SaveUser(ctx, seed)
	if err != nil {
		log.Fatal(err)
	}

	// Match the TS example: delete users where username == "obsolete".
	deletedCount, err := db.DeleteUsers(ctx).Where(onyx.Eq("username", "obsolete")).Delete()
	if err != nil {
		log.Fatal(err)
	}
	if deletedCount == 0 {
		log.Println("warning: expected to delete seeded user")
	}

	fmt.Printf("Deleted %d record(s).\n", deletedCount)
	log.Println("example: completed")
}
