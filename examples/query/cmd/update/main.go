package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	updated, err := db.From("User").
		Where(contract.Eq("id", "example-user-1")).
		SetUpdates(map[string]any{"isActive": false}).
		Update(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Updated %d record(s).\n", updated)
}
