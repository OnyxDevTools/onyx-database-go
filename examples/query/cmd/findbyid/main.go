package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/gen/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	id := "example-user-1"
	results, err := db.Users().
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
