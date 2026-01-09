package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyxclient"
)

func main() {
	ctx := context.Background()
	streamCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := onyxclient.NewClient(core)

	iter, err := db.Users(streamCtx).Stream(streamCtx)
	if err != nil {
		log.Fatal(err)
	}

	// Trigger at least one event so the stream returns promptly.
	go func() {
		time.Sleep(200 * time.Millisecond)
		now := time.Now().UTC()
		_, _ = db.Users(ctx).Save(onyxclient.User{
			Id:        "stream_query_user",
			Username:  "stream-query",
			Email:     "stream-query@example.com",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}()
	if iter == nil {
		log.Fatalf("warning: expected stream iterator")
		return
	}
	defer func() {
		if err := iter.Close(); err != nil {
			log.Printf("stream close error: %v", err)
		}
	}()

	count := 0
	for iter.Next() {
		user := iter.Value()
		if user == nil {
			log.Fatalf("warning: expected streamed user")
		}
		fmt.Println("USER:", user)
		count++
		if count >= 3 {
			break
		}
	}

	if err := iter.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		log.Fatal(err)
	}
	log.Println("example: completed")
}
