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

	users, err := db.Users().
		Select("id", "email", "createdAt").
		OrderBy("createdAt", false).
		Limit(3).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if users == nil {
		log.Fatalf("warning: expected users response")
	}

	out, _ := json.MarshalIndent(users, "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}
