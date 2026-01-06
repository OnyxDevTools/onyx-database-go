package main

import (
	"context"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	streamDB, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	writeDB, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	iter, err := streamDB.From("User").Stream(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer iter.Close()

	go func() {
		time.Sleep(200 * time.Millisecond)
		_, _ = writeDB.Save(ctx, "User", map[string]any{
			"id":        "stream_user_create",
			"username":  "create-user",
			"email":     "create@example.com",
			"isActive":  true,
			"createdAt": time.Now().UTC().Format(time.RFC3339),
			"updatedAt": time.Now().UTC().Format(time.RFC3339),
		}, nil)
	}()

	for iter.Next() {
		log.Printf("USER CREATED: %+v", iter.Value())
		break
	}
}
