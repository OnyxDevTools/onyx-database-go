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

	targetID := "user-id-1"
	if err := db.Delete(ctx, "User", targetID); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deleted 1 record(s).")
}
