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

	users, err := db.From("User").
		Where(contract.Eq("username", "obsolete")).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	deleted := 0
	for _, user := range users {
		id, ok := user["id"].(string)
		if !ok {
			continue
		}
		if err := db.Delete(ctx, "User", id); err == nil {
			deleted++
		}
	}

	fmt.Printf("Deleted %d record(s).\n", deleted)
}
