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
	docs := db.Documents()

	doc := contract.Document{
		ID: "note.json",
		Data: map[string]any{
			"message": "hello",
		},
	}

	saved, err := docs.Save(ctx, doc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("saved document: %+v\n", saved)

	found, err := docs.Get(ctx, saved.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Document contents: %+v\n", found)

	if err := docs.Delete(ctx, saved.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted document:", saved.ID)

	if _, err := docs.Get(ctx, saved.ID); err != nil {
		fmt.Println("Document was deleted successfully")
	} else {
		fmt.Println("oops document still exists")
	}
}
