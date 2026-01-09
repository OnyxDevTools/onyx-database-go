package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyxdb"
)

func main() {
	ctx := context.Background()

	db, err := onyxdb.New(ctx, onyxdb.Config{})
	if err != nil {
		log.Fatal(err)
	}

	logs, err := db.AuditLogs().
		Select("actorId", "action", "targetId", "status", "dateTime").
		Where(onyxdb.Eq("actorId", "admin-user-1")).
		And(onyxdb.Eq("action", "DELETE")).
		Or(onyxdb.Eq("action", "UPDATE")).
		Or(onyxdb.NotNull("actorId")).
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
