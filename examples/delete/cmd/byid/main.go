package main

import (
	"context"
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

	targetID := "user-id-1"
	now := time.Now().UTC()
	saved, err := db.SaveUser(ctx, onyx.User{
		Id:        targetID,
		Asdf:      "tmp",
		Username:  "delete_me",
		Email:     "delete_me@example.com",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		log.Fatal(err)
	}
	if saved.Id == "" {
		log.Fatalf("warning: expected saved user id")
	}

	fmt.Printf("Saved user: %+v\n", saved)

	deleted, err := db.DeleteUser(ctx, targetID)
	if err != nil {
		log.Fatal(err)
	}
	if deleted == 0 {
		log.Fatalf("warning: expected to delete saved user")
	}
	fmt.Printf("Deleted %d record(s).\n", deleted)

	fmt.Println("Done.")
	log.Println("example: completed")
}
