package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	id := "example-user-1"
	results, err := db.From("User").Where(contract.Eq("id", id)).Limit(1).List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if len(results) == 0 {
		fmt.Println("No record found for id:", id)
		return
	}

	out, _ := json.MarshalIndent(results[0], "", "  ")
	fmt.Println(string(out))
}
