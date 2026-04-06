package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
	coreonyx "github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	stats, err := db.UserProfiles().
		Select(
			coreonyx.Sum("age"),
			coreonyx.Min("age"),
			coreonyx.Max("age"),
			coreonyx.Median("age"),
			coreonyx.Percentile("age", 95),
			coreonyx.Std("age"),
			coreonyx.Variance("age"),
		).
		List(ctx)
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
