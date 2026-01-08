package main

import (
	"context"
	"encoding/json"
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

	users, err := db.ListUsers().
		Select("username", "email").
		Limit(2).
		ListMaps(ctx)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(users, "", "  ")
	fmt.Println(string(out))
}
