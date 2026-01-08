package main

import (
	"context"
	"encoding/json"
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

	first, err := db.ListUsers().
		Where(onyx.Eq("email", "basic@example.com")).
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

	also, err := db.ListUsers().
		Where(onyx.Eq("email", "notfound@example.com")).
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
