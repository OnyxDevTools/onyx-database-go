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

	id := "first-or-null-user-1"
	now := time.Now()

	// Seed a user so there is always data to read.
	seed := onyx.User{
		Id:          id,
		Username:    "First Or Null User",
		Email:       "first-or-null@example.com",
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
		LastLoginAt: nil,
	}
	if _, err := db.Users().Save(ctx, seed); err != nil {
		log.Fatalf("failed to seed user: %v", err)
	}

	// Fetch the newest user by createdAt without adding a where() clause.
	latest, err := db.Users().
		OrderBy("createdAt", false).
		Limit(1).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if len(latest) == 0 {
		log.Fatal("expected a record from Limit+OrderBy without where")
	}

	out, _ := json.MarshalIndent(latest[0], "", "  ")
	fmt.Printf("Latest user: %s\n", string(out))

	log.Println("example: completed")
}
