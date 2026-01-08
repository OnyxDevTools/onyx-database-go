package main

import (
	"context"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	original, err := db.GetSchema(ctx, nil)
	if err != nil {
		log.Fatalf("failed to fetch schema: %v", err)
	}

	if len(original.Tables) == 0 {
		log.Fatal("expected non-empty schema")
	}

	temp := onyx.Table{
		Name: "TempTable",
		Fields: []onyx.Field{
			{Name: "id", Type: "String", Primary: true},
			{Name: "name", Type: "String"},
		},
	}

	withTemp := addTable(original, temp)
	if err := db.ValidateSchema(ctx, withTemp); err != nil {
		log.Fatalf("schema validation failed: %v", err)
	}
	if !hasTable(original, temp.Name) && hasTable(withTemp, temp.Name) {
		log.Printf("diff: %s added", temp.Name)
	}

	if err := db.PublishSchema(ctx, withTemp); err != nil {
		log.Fatalf("publish with temp failed: %v", err)
	}
	published, err := db.GetSchema(ctx, nil)
	if err != nil {
		log.Fatalf("failed to fetch schema after publish: %v", err)
	}
	if !hasTable(published, temp.Name) {
		log.Fatalf("expected %s to be present after publish", temp.Name)
	}
	log.Printf("%s added and published", temp.Name)

	withoutTemp := removeTable(published, temp.Name)
	if err := db.ValidateSchema(ctx, withoutTemp); err != nil {
		log.Fatalf("schema validation (remove temp) failed: %v", err)
	}
	if err := db.PublishSchema(ctx, withoutTemp); err != nil {
		log.Fatalf("publish without temp failed: %v", err)
	}
	finalSchema, err := db.GetSchema(ctx, nil)
	if err != nil {
		log.Fatalf("failed to fetch schema after removal: %v", err)
	}
	if hasTable(finalSchema, temp.Name) {
		log.Fatalf("expected %s to be removed after publish", temp.Name)
	}
	log.Printf("all operations worked as expected, %s added, removed and published", temp.Name)
}

func hasTable(schema onyx.Schema, name string) bool {
	for _, t := range schema.Tables {
		if t.Name == name {
			return true
		}
	}
	return false
}

func addTable(schema onyx.Schema, table onyx.Table) onyx.Schema {
	if hasTable(schema, table.Name) {
		return schema
	}
	copyTables := append([]onyx.Table{}, schema.Tables...)
	copyTables = append(copyTables, table)
	return onyx.Schema{Tables: copyTables}
}

func removeTable(schema onyx.Schema, name string) onyx.Schema {
	out := onyx.Schema{}
	for _, t := range schema.Tables {
		if t.Name == name {
			continue
		}
		out.Tables = append(out.Tables, t)
	}
	return out
}
