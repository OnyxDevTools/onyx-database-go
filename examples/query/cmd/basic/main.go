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

	users, err := db.ListUsers().
		Where(onyx.Eq("isActive", true)).
		Limit(5).
		List(ctx)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	if users == nil {
		log.Fatalf("warning: expected users response")
	}

	pretty, _ := json.MarshalIndent(users, "", "  ")
	fmt.Println(string(pretty))
	log.Println("example: completed")
}
