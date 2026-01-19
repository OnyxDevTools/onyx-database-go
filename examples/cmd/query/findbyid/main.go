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

	id := "example-user-1"
	user, err := db.Users().FindByID(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	if user.Id == "" {
		fmt.Println("No record found for id:", id)
		log.Println("example: completed")
		return
	}

	out, _ := json.MarshalIndent(user, "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}
