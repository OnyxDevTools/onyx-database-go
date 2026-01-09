package main

import (
	"context"
	"fmt"
	"log"

	model "github.com/OnyxDevTools/onyx-database-go/examples/onyx"
)

func main() {
	ctx := context.Background()

	db, err := model.New(ctx, model.Config{})
	if err != nil {
		log.Fatal(err)
	}

	users, err := db.Users().Limit(5).List(ctx)
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
