package main

import (
	"context"
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

	q := db.From("User").OrderBy(contract.Desc("username")).Limit(2)

	firstPage, err := q.Page(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Page 1:", usernames(firstPage.Items))

	if firstPage.NextCursor != "" {
		secondPage, err := q.Page(ctx, firstPage.NextCursor)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Page 2:", usernames(secondPage.Items))
	}
}

func usernames(items contract.QueryResults) []any {
	var names []any
	for _, u := range items {
		if name, ok := u["username"]; ok {
			names = append(names, name)
		}
	}
	return names
}
