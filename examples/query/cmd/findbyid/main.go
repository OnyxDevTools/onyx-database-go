package main

import (
	"context"
	"encoding/json"
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

	id := "example-user-1"
	results, err := db.ListUsers().
		Where(onyx.Eq("id", id)).
		Limit(1).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if results == nil {
		log.Fatalf("warning: expected query response")
	}

	if len(results) == 0 {
		fmt.Println("No record found for id:", id)
		return
	}
	if results[0].Id == "" {
		log.Fatalf("warning: expected user id")
	}

	out, _ := json.MarshalIndent(results[0], "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}
