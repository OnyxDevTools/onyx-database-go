package main

import (
	"context"
	"fmt"
	"log"

	model "github.com/OnyxDevTools/onyx-database-go/examples/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := model.NewClient(core)

	users, err := db.ListUsers().Limit(5).List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if users == nil {
		log.Fatalf("warning: expected users response")
	}

	for _, u := range users {
		fmt.Println(u.Username)
	}
	log.Println("example: completed")
}
