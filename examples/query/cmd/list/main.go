package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	users, err := db.From("User").Limit(5).List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		if name, ok := u["username"]; ok {
			fmt.Println(name)
		}
	}
}
