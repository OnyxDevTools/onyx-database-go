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
	db := onyxclient.NewClient(core)

	now := time.Now().UTC()
	user := onyxclient.User{
		Id:        "example-user-1", // if you omit this one will be generated when the schema has a UUID generator specified
		Username:  "Example User",
		Email:     "basic@example.com",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	saved, err := db.SaveUser(ctx, user)
	if err != nil {
		log.Fatal(err)
	}

	jsonOut, _ := json.Marshal(saved)
	fmt.Printf("Saved user: %s\n", string(jsonOut))
}
