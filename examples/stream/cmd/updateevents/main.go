package main

import (
	"context"
	"log"
	"time"

	model "github.com/OnyxDevTools/onyx-database-go/examples/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	streamDB, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	writeDB, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	iter, err := streamDB.From(model.Tables.User).Stream(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer iter.Close()

	go func() {
		time.Sleep(200 * time.Millisecond)
		now := time.Now().UTC()
		_, _ = writeDB.Save(ctx, model.Tables.User, model.User{
			Id:        "stream_user_update",
			Username:  "update-user",
			Email:     "update@example.com",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil)
		time.Sleep(200 * time.Millisecond)
		updated := time.Now().UTC()
		lastLogin := updated
		_, _ = writeDB.Save(ctx, model.Tables.User, model.User{
			Id:          "stream_user_update",
			Username:    "update-user-updated",
			Email:       "update@example.com",
			IsActive:    true,
			LastLoginAt: &lastLogin,
			CreatedAt:   updated,
			UpdatedAt:   updated,
		}, nil)
	}()

	for iter.Next() {
		log.Printf("USER EVENT: %+v", iter.Value())
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
}
