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

	stats, err := db.ListUserProfiles().
		Select("avg(age)").
		ListAggregates(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if stats == nil {
		log.Fatalf("warning: expected aggregate stats")
	}

	out, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}
