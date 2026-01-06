package main

import (
	"context"
	"encoding/json"
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

	logs, err := db.From("AuditLog").
		Select("actorId", "action", "targetId", "status", "dateTime").
		Where(contract.Eq("actorId", "admin-user-1")).
		And(contract.Eq("action", "DELETE")).
		Or(contract.Eq("action", "UPDATE")).
		Or(contract.NotNull("actorId")).
		OrderBy(contract.Desc("dateTime")).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(logs, "", "  ")
	fmt.Println(string(out))
}
