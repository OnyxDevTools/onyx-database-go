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

	users, err := db.Users(ctx).
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
