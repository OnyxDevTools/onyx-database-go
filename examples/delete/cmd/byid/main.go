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

	targetID := "user-id-1"
	now := time.Now().UTC()
	saved, err := db.Users(ctx).Save(onyxclient.User{
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

	if err := db.Users(ctx).DeleteByID(targetID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted saved user.")

	fmt.Println("Done.")
	log.Println("example: completed")
}
