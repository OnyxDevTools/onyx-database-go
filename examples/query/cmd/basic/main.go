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

	results, err := db.From("User").
		Where(contract.Eq("isActive", true)).
		Limit(5).
		List(ctx)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}

	pretty, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(pretty))
}
