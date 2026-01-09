package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()
	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := core.Typed()

	iter, err := db.ListUsers().Stream(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if iter == nil {
		log.Fatalf("warning: expected stream iterator")
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
	if err := iter.Err(); err != nil &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, context.DeadlineExceeded) &&
		err.Error() != "http2: response body closed" {
		log.Fatal(err)
	}
	log.Println("example: completed")
}
