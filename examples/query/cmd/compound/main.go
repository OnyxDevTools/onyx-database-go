package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := core.Typed()

	logs, err := db.ListAuditLogs().
		Select("actorId", "action", "targetId", "status", "dateTime").
		Where(onyx.Eq("actorId", "admin-user-1")).
		And(onyx.Eq("action", "DELETE")).
		Or(onyx.Eq("action", "UPDATE")).
		Or(onyx.NotNull("actorId")).
		OrderBy("dateTime", false).
		ListMaps(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if logs == nil {
		log.Fatalf("warning: expected audit logs response")
	}

	out, _ := json.MarshalIndent(logs, "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}
