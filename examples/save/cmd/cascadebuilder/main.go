package main

import (
	"context"
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

	spec := contract.NewCascadeBuilder().
		Graph("profile").
		GraphType("UserProfile").
		TargetField("userId").
		SourceField("id").
		Build()

	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	user := map[string]any{
		"id":        "cb-user-1",
		"username":  "Cascade Builder",
		"email":     "cascade-builder@example.com",
		"isActive":  true,
		"createdAt": now,
		"updatedAt": now,
		"profile": map[string]any{
			"id":        "cb-profile-1",
			"userId":    "cb-user-1",
			"firstName": "Cascade",
			"lastName":  "Builder",
		},
	}

	if err := db.Cascade(spec).Save(ctx, "User", user); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Saved user with cascadeBuilder")
}
