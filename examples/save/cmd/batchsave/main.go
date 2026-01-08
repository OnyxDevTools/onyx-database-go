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

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	client := onyxclient.NewClient(db)

	now := time.Now().UTC()
	users := make([]onyxclient.User, 0, 5)
	for i := 0; i < 5; i++ {
		users = append(users, onyxclient.User{
			Id:        fmt.Sprintf("batch-user-%d", i),
			Username:  fmt.Sprintf("Batch User %d", i),
			Email:     fmt.Sprintf("batch%d@example.com", i),
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if len(users) == 0 {
		log.Println("warning: expected users to save")
	}
	if err := db.BatchSave(ctx, onyxclient.Tables.User, toAnySlice(users), 2); err != nil {
		log.Fatal(err)
	}

	savedCount := len(users)
	fmt.Printf("Batch saved users: %d\n", savedCount)

	// Fetch a small sample with a timeout so debug sessions don't hang if the network is slow.
	listCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if fetched, err := client.ListUsers().Limit(5).List(listCtx); err == nil {
		var decoded []onyxclient.User
		if b, marshalErr := json.Marshal(fetched); marshalErr == nil {
			_ = json.Unmarshal(b, &decoded)
		}
		fmt.Printf("Fetched %d users (first 5):\n", len(decoded))
		for _, u := range decoded {
			fmt.Printf("- %s\n", u.Id)
		}
	} else {
		fmt.Printf("Fetch skipped: %v\n", err)
	}
	log.Println("example: completed")
}

func toAnySlice[T any](in []T) []any {
	out := make([]any, len(in))
	for i := range in {
		out[i] = in[i]
	}
	return out
}
