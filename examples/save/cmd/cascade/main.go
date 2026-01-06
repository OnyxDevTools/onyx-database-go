package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	userID := "example_user2"
	spec := contract.Cascade("profile:UserProfile(userId,id)")

	cascade := db.Cascade(spec)
	user := map[string]any{
		"id":        userID,
		"username":  "cascade",
		"email":     "cascade@example.com",
		"isActive":  true,
		"createdAt": now,
		"updatedAt": now,
		"profile": map[string]any{
			"id":        "profile_001",
			"userId":    userID,
			"firstName": "Test",
			"lastName":  "User",
			"age":       24,
			"createdAt": now,
			"updatedAt": now,
		},
	}

	if err := cascade.Save(ctx, "User", user); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Saved user with cascade.")

	users, err := db.From("User").
		Where(contract.Eq("id", userID)).
		Resolve("profile").
		Limit(1).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(users, "", "  ")
	fmt.Printf("retrieved user with profile: %s\n", string(out))
}
