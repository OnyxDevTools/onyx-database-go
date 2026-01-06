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
	users := make([]any, 0, 5)
	for i := 0; i < 5; i++ {
		users = append(users, map[string]any{
			"id":        fmt.Sprintf("batch-user-%d", i),
			"username":  fmt.Sprintf("Batch User %d", i),
			"email":     fmt.Sprintf("batch%d@example.com", i),
			"isActive":  true,
			"createdAt": now,
			"updatedAt": now,
		})
	}

	if err := db.BatchSave(ctx, "User", users, 2); err != nil {
		log.Fatal(err)
	}

	body, _ := json.Marshal(map[string]any{"saved": len(users)})
	fmt.Printf("Batch saved users: %d\n", len(users))
	fmt.Println(string(body))
}
