package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/gen/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	q := db.Users().OrderBy("username", false).Limit(2)

	firstPage, err := q.Page(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	if firstPage.Items == nil {
		log.Fatalf("warning: expected page items")
	}
	fmt.Println("Page 1:", usernames(firstPage.Items))

	if firstPage.NextCursor != "" {
		secondPage, err := q.Page(ctx, firstPage.NextCursor)
		if err != nil {
			log.Fatal(err)
		}
		if secondPage.Items == nil {
			log.Fatalf("warning: expected second page items")
		}
		fmt.Println("Page 2:", usernames(secondPage.Items))
	}
	log.Println("example: completed")
}

func usernames(items []onyx.User) []string {
	var names []string
	for _, u := range items {
		names = append(names, u.Username)
	}
	return names
}
