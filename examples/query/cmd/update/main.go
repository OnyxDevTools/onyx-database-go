package main

import (
	"context"
	"fmt"
	"log"

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

	updated, err := db.Users(ctx).
		Where(onyx.Eq("id", "example-user-1")).
		SetUpdates(map[string]any{"isActive": false}).
		Update(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if updated < 0 {
		log.Fatalf("warning: expected update count")
	}

	fmt.Printf("Updated %d record(s).\n", updated)
	log.Println("example: completed")
}
