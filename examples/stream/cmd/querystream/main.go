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

	iter, err := db.From("User").Stream(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer iter.Close()

	for iter.Next() {
		user := iter.Value()
		fmt.Println("USER:", user)
	}

	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
}
