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

	client, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now().UTC()
	userID := "example_user2"
	user := onyx.User{
		Id:        userID,
		Username:  "cascade",
		Email:     "cascade@example.com",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		Profile: onyx.UserProfile{
			Id:        "profile_001",
			UserId:    userID,
			FirstName: "Test",
			LastName:  "User",
			Age:       int64Ptr(24),
			CreatedAt: now,
			UpdatedAt: &now,
		},
	}

	// Cascade save using a CascadeSpec; Users().Save returns the saved graph.
	spec := onyx.Cascade("profile:UserProfile(userId,id)")
	saved, err := client.Users().Save(ctx, user, spec)
	if err != nil {
		log.Fatal(err)
	}
	if saved.Id == "" {
		log.Fatalf("warning: expected saved user id")
	}
	fmt.Println("Saved user with cascade.")

	outUser, _ := json.MarshalIndent(saved, "", "  ")
	fmt.Printf("saved user: %s\n", string(outUser))
	if saved.Profile != nil {
		outProfile, _ := json.MarshalIndent(saved.Profile, "", "  ")
		fmt.Printf("saved profile: %s\n", string(outProfile))
	} else {
		log.Fatalf("warning: expected profile to be returned in save response")
	}
	log.Println("example: completed")
}

func int64Ptr(v int64) *int64 {
	return &v
}
