package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now().UTC()
	user := onyx.User{
		Id:        "example-user-1", // if you omit this one will be generated when the schema has a UUID generator specified
		Username:  "Example User",
		Email:     "basic@example.com",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	saved, err := db.Users().Save(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	if saved.Id == "" {
		log.Fatalf("warning: expected saved user id")
	}

	jsonOut, _ := json.Marshal(saved)
	fmt.Printf("Saved user: %s\n", string(jsonOut))
	log.Println("example: completed")
}
