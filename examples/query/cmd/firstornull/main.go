package main

import (
	"context"
	"encoding/json"
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

	first, err := db.From("User").
		Where(contract.Eq("email", "basic@example.com")).
		Limit(1).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if len(first) > 0 {
		out, _ := json.MarshalIndent(first[0], "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Println("null")
	}

	also, err := db.From("User").
		Where(contract.Eq("email", "notfound@example.com")).
		Limit(1).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if len(also) == 0 {
		fmt.Println("\nshould be null: null")
	} else {
		out, _ := json.MarshalIndent(also[0], "", "  ")
		fmt.Printf("\nshould be null: %s\n", string(out))
	}
}
