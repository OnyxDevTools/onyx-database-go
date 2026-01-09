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

	stats, err := db.ListUsers().
		Select("isActive", "count(id)").
		GroupBy("isActive").
		ListAggregates(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if stats == nil {
		log.Fatalf("warning: expected grouped aggregates")
	}

	out, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}
