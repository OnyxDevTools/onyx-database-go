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

	users, err := db.Users().
		Where(onyxdb.Eq("isActive", true)).
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
