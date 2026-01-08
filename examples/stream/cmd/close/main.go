package main

import (
	"context"
	"log"
	"time"

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
	if iter == nil {
		log.Println("warning: expected stream iterator")
		return
	}
	defer func() {
		if err := iter.Close(); err != nil {
			log.Printf("stream close error: %v", err)
		}
	}()

	// Let the stream start, then cancel shortly after to demonstrate cleanup.
	time.AfterFunc(500*time.Millisecond, func() {
		_ = iter.Close()
	})

	for iter.Next() {
		// ignore records; this example just shows closing the stream.
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println("example: completed")
}
