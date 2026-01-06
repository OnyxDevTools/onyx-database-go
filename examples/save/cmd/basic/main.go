package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	user := map[string]any{
		"id":        "example-user-1", // if you omit this one will be generated for you when the schema has a UUID generator specified
		"username":  "Example User",
		"email":     "basic@example.com",
		"isActive":  true,
		"createdAt": now,
		"updatedAt": now,
	}

	saved, err := db.Save(ctx, "User", user, nil)
	if err != nil {
		log.Fatal(err)
	}

	jsonOut, _ := json.Marshal(saved)
	fmt.Printf("Saved user: %s\n", string(jsonOut))
}
