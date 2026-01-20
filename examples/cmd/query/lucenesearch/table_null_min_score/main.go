package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	gen "github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
	sdk "github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := gen.New(ctx, gen.Config{})
	if err != nil {
		log.Fatal(err)
	}

	searchText := "Text"

	now := time.Now().UTC()
	id := newID()

	saved, err := db.Users().Save(ctx, gen.User{
		Id:        id,
		Username:  "Lucene Text target (table, null minScore)",
		Email:     fmt.Sprintf("lucene-table-nullscore-%d@example.com", time.Now().UnixNano()),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		log.Fatalf("FAILED: failed to seed user: %v", err)
	}
	if saved.Id == "" {
		log.Fatalf("FAILED: server did not return an id for saved user")
	}

	found := false
	for attempt := 1; attempt <= 6; attempt++ {
		users, err := db.Users().
			Where(sdk.Search(searchText)).
			List(ctx)
		if err != nil {
			log.Fatalf("FAILED: search failed: %v", err)
		}
		for _, u := range users {
			if u.Id == saved.Id {
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if !found {
		log.Fatalf("FAILED: expected user %s in search results", saved.Id)
	}

	log.Println("found seeded user in table search with null minScore!")
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}
