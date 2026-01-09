package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/gen/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	targetID := "user-id-1"
	now := time.Now().UTC()
	saved, err := db.Users().Save(ctx, onyx.User{
		Id:        targetID,
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

	if _, err := db.Users().DeleteByID(ctx, targetID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted saved user.")

	fmt.Println("Done.")
	log.Println("example: completed")
}
