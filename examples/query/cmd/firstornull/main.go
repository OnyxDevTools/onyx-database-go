package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyxdb"
)

func main() {
	ctx := context.Background()

	db, err := onyxdb.New(ctx, onyxdb.Config{})
	if err != nil {
		log.Fatal(err)
	}

	first, err := db.Users().
		Where(onyxdb.Eq("email", "basic@example.com")).
		Limit(1).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if first == nil {
		log.Fatalf("warning: expected query response")
	}

	if len(first) > 0 {
		if first[0].Id == "" {
			log.Fatalf("warning: expected user id")
		}
		out, _ := json.MarshalIndent(first[0], "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Println("null")
	}

	also, err := db.Users().
		Where(onyxdb.Eq("email", "notfound@example.com")).
		Limit(1).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if also == nil {
		log.Fatalf("warning: expected query response")
	}

	if len(also) == 0 {
		fmt.Println("\nshould be null: null")
	} else {
		if also[0].Id == "" {
			log.Fatalf("warning: expected user id")
		}
		out, _ := json.MarshalIndent(also[0], "", "  ")
		fmt.Printf("\nshould be null: %s\n", string(out))
	}
	log.Println("example: completed")
}
