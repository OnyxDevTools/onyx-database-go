package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
)

func main() {
	ctx := context.Background()
	streamCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	streamDB, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	writeDB, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	iter, err := streamDB.Users().Stream(streamCtx)
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

	go func() {
		time.Sleep(200 * time.Millisecond)
		now := time.Now().UTC()
		_, err := writeDB.Users().Save(ctx, onyx.User{
			Id:        "stream_user_create",
			Username:  "create-user",
			Email:     "create@example.com",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			log.Printf("save error: %v", err)
		}
	}()

	if iter.Next() {
		if iter.Value() == nil {
			log.Fatalf("warning: expected streamed value")
		}
		fmt.Printf("USER CREATED: %+v\n", iter.Value())
	}
	if err := iter.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		log.Fatal(err)
	}
	log.Println("example: completed")
}
