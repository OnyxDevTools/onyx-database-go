package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	logs, err := db.AuditLogs().
		Select("actorId", "action", "targetId", "status", "dateTime").
		Where(onyx.Eq("actorId", "admin-user-1")).
		And(onyx.Eq("action", "DELETE")).
		Or(onyx.Eq("action", "UPDATE")).
		Or(onyx.NotNull("actorId")).
		OrderBy("dateTime", false).
		List(ctx)

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
