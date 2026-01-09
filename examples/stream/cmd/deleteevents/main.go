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

	streamCore, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	writeCore, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	streamDB := onyxclient.NewClient(streamCore)
	writeDB := onyxclient.NewClient(writeCore)

	iter, err := streamDB.Users(streamCtx).Stream(streamCtx)
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

	// seed then delete to trigger an event
	go func() {
		time.Sleep(200 * time.Millisecond)
		now := time.Now().UTC()
		_, err := writeDB.Users(ctx).Save(onyxclient.User{
			Id:        "stream_user_delete",
			Username:  "delete-user",
			Email:     "delete@example.com",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			log.Printf("save error: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
		if err := writeDB.Users(ctx).DeleteByID("stream_user_delete"); err != nil {
			log.Printf("delete error: %v", err)
		}
	}()

	if iter.Next() {
		if iter.Value() == nil {
			log.Fatalf("warning: expected streamed value")
		}
		fmt.Printf("USER EVENT: %+v\n", iter.Value())
	}
	if err := iter.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		log.Fatal(err)
	}
	log.Println("example: completed")
}
