package main

import (
	"context"
	"fmt"
	"log"
	"time"

	model "github.com/OnyxDevTools/onyx-database-go/examples/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()
	streamCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	streamDB, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	writeDB, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	iter, err := streamDB.From(model.Tables.User).Stream(streamCtx)
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

	go func() {
		time.Sleep(200 * time.Millisecond)
		now := time.Now().UTC()
		_, err := writeDB.Save(ctx, model.Tables.User, model.User{
			Id:        "stream_user_update",
			Username:  "update-user",
			Email:     "update@example.com",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil)
		if err != nil {
			log.Printf("save error: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
		updated := time.Now().UTC()
		lastLogin := updated
		_, err = writeDB.Save(ctx, model.Tables.User, model.User{
			Id:          "stream_user_update",
			Username:    "update-user-updated",
			Email:       "update@example.com",
			IsActive:    true,
			LastLoginAt: &lastLogin,
			CreatedAt:   updated,
			UpdatedAt:   updated,
		}, nil)
		if err != nil {
			log.Printf("save error: %v", err)
		}
	}()

	eventCount := 0
	for iter.Next() {
		if iter.Value() == nil {
			log.Println("warning: expected streamed value")
		}
		fmt.Printf("USER EVENT: %+v\n", iter.Value())
		eventCount++
		if eventCount >= 2 {
			break
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println("example: completed")
}
