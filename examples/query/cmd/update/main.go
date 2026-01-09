package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyxdb"
)

func main() {
	ctx := context.Background()

	db, err := onyxdb.New(ctx, onyxdb.Config{})
	if err != nil {
		log.Fatal(err)
	}

	updated, err := db.Users().
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
