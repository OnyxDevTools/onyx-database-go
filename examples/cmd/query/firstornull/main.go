package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Exactly one match expected.
	first, err := db.Users().
		Where(onyx.Eq("email", "basic@example.com")).
		Limit(1).
		One(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if first.Username == "" {
		log.Fatalf("warning: expected a single User object")
	}
	out, _ := json.MarshalIndent(first, "", "  ")
	fmt.Println(string(out))

	// Zero matches expected.
	also, err := db.Users().
		Where(onyx.Eq("email", "notfound@example.com")).
		Limit(1).
		FirstOrNil(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if also == nil {
		fmt.Println("\nshould be null: null")
	} else if also.Username == "" {
		log.Fatalf("warning: expected a User object when not null")
	} else {
		out, _ := json.MarshalIndent(also, "", "  ")
		fmt.Printf("\nshould be null: %s\n", string(out))
	}
	log.Println("example: completed")
}
