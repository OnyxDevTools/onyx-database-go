package main

import (
	"context"
	"encoding/json"
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

	client := onyxclient.NewClient(core)

	now := time.Now().UTC()
	userID := "example_user2"
	user := onyxclient.User{
		Id:        userID,
		Username:  "cascade",
		Email:     "cascade@example.com",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		Profile: onyxclient.UserProfile{
			Id:        "profile_001",
			UserId:    userID,
			FirstName: "Test",
			LastName:  "User",
			Age:       int64Ptr(24),
			CreatedAt: now,
			UpdatedAt: &now,
		},
	}

	// Cascade save using a CascadeSpec; SaveUser returns the saved graph.
	spec := onyx.Cascade("profile:UserProfile(userId,id)")
	saved, err := client.SaveUser(ctx, user, spec)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Saved user with cascade.")

	outUser, _ := json.MarshalIndent(saved, "", "  ")
	fmt.Printf("saved user: %s\n", string(outUser))
	if saved.Profile != nil {
		outProfile, _ := json.MarshalIndent(saved.Profile, "", "  ")
		fmt.Printf("saved profile: %s\n", string(outProfile))
	} else {
		log.Fatal("expected profile to be returned in save response")
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
