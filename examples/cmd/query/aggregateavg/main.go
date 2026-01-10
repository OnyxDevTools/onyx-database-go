package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	stats, err := db.UserProfiles().
		Select("avg(age)").
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
