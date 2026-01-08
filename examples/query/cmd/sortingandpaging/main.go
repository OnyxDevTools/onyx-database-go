package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyxclient"
)

func main() {
	ctx := context.Background()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := onyxclient.NewClient(core)

	q := db.ListUsers().OrderBy("username", false).Limit(2)

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

func usernames(items []onyxclient.User) []string {
	var names []string
	for _, u := range items {
		names = append(names, u.Username)
	}
	return names
}
