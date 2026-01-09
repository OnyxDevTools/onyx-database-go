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
		Select("id", "email", "createdAt").
		OrderBy("createdAt", false).
		Limit(3).
		ListMaps(ctx)
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
