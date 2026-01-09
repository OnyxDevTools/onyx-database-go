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

	spec := onyx.NewCascadeBuilder().
		Graph("profile").
		GraphType("UserProfile").
		SourceField("userId").
		TargetField("id").
		Build()

	now := time.Now().UTC()
	user := onyxclient.User{
		Id:        "cb-user-1",
		Username:  "Cascade Builder",
		Email:     "cascade-builder@example.com",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		Profile: onyxclient.UserProfile{
			Id:        "cb-profile-1",
			UserId:    "cb-user-1",
			FirstName: "Cascade",
			LastName:  "Builder",
		},
	}

	saved, err := db.SaveUser(ctx, user, spec)
	if err != nil {
		log.Fatal(err)
	}
	if saved.Id == "" {
		log.Fatalf("warning: expected saved user id")
	}
	if saved.Profile == nil {
		log.Fatalf("warning: expected saved profile")
	}
	fmt.Println("Saved user with cascadeBuilder")
	log.Println("example: completed")
}
