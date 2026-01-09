package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := core.Typed()

	updated, err := db.ListUsers().
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
