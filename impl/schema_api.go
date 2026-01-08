package impl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type schemaEntity struct {
	Name       string            `json:"name"`
	Identifier *schemaIdentifier `json:"identifier,omitempty"`
	Attributes []schemaAttribute `json:"attributes,omitempty"`
	Partition  string            `json:"partition,omitempty"`
	Indexes    []map[string]any  `json:"indexes,omitempty"`
	Resolvers  []map[string]any  `json:"resolvers,omitempty"`
	Triggers   []map[string]any  `json:"triggers,omitempty"`
	Meta       map[string]any    `json:"meta,omitempty"`
}

type schemaIdentifier struct {
	Name      string `json:"name"`
	Generator string `json:"generator,omitempty"`
	Type      string `json:"type,omitempty"`
}

type schemaAttribute struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsNullable bool   `json:"isNullable,omitempty"`
}

func fetchSchema(ctx context.Context, c *client, tables []string) (contract.Schema, error) {
	var raw map[string]any
	params := url.Values{}
	if len(tables) > 0 {
		params.Set("tables", strings.Join(tables, ","))
	}
	path := "/database/" + tableEscape(c.cfg.DatabaseID) + "/schema"
	if enc := params.Encode(); enc != "" {
		path += "?" + enc
	}
	err := c.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &raw)
	if err != nil {
		if cerr, ok := err.(*contract.Error); ok {
			if status, ok := cerr.Meta["status"].(int); ok && status == http.StatusNotFound {
				// Try /schemas/{db} (history endpoint) then legacy /schema.
				schemaPath := "/schemas/" + tableEscape(c.cfg.DatabaseID)
				if err2 := c.httpClient.DoJSON(ctx, http.MethodGet, schemaPath, nil, &raw); err2 == nil {
					// raw may be a list or object; parsing handled below.
				} else if err3 := c.httpClient.DoJSON(ctx, http.MethodGet, "/schema", nil, &raw); err3 != nil {
					return contract.Schema{}, err3
				}
			} else {
				return contract.Schema{}, err
			}
		} else {
			return contract.Schema{}, err
		}
	}
	if cleaned, ok := stripEntityText(raw).(map[string]any); ok {
		raw = cleaned
	}
	if list, ok := raw["schemas"].([]any); ok && len(list) > 0 {
		// take latest entry
		if obj, ok := list[len(list)-1].(map[string]any); ok {
			return schemaFromEntities(obj["entities"].([]any)), nil
		}
	}
	if schemaObj, ok := raw["schema"].(map[string]any); ok {
		if entities, ok := schemaObj["entities"].([]any); ok {
			return schemaFromEntities(entities), nil
		}
		if tableSlice, ok := schemaObj["tables"].([]any); ok {
			return schemaFromTablesArray(tableSlice), nil
		}
	}
	if entities, ok := raw["entities"].([]any); ok {
		return schemaFromEntities(entities), nil
	}
	if tableSlice, ok := raw["tables"].([]any); ok {
		return schemaFromTablesArray(tableSlice), nil
	}
	if tablesMap, ok := raw["tables"].(map[string]any); ok {
		tables := make([]map[string]any, 0, len(tablesMap))
		for name, val := range tablesMap {
			table := map[string]any{"name": name}
			if fieldsMap, ok := val.(map[string]any)["fields"].(map[string]any); ok {
				fields := make([]map[string]any, 0, len(fieldsMap))
				for fieldName, fval := range fieldsMap {
					fieldObj := map[string]any{"name": fieldName}
					if fattrs, ok := fval.(map[string]any); ok {
						for k, v := range fattrs {
							fieldObj[k] = v
						}
					}
					fields = append(fields, fieldObj)
				}
				table["fields"] = fields
			}
			tables = append(tables, table)
		}
		raw["tables"] = tables
	}
	data, err := jsonMarshal(raw)
	if err != nil {
		return contract.Schema{}, err
	}
	return contract.ParseSchemaJSON(data)
}

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// publishSchema sends the schema using the TS-style endpoint /schemas/{databaseId}.
func publishSchema(ctx context.Context, c *client, schema contract.Schema, publish bool) error {
	normalized := contract.NormalizeSchema(schema)
	params := url.Values{}
	if publish {
		params.Set("publish", "true")
	}
	path := "/schemas/" + tableEscape(c.cfg.DatabaseID)
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return c.httpClient.DoJSON(ctx, http.MethodPut, path, schemaUpsertPayload(normalized, c.cfg.DatabaseID), nil)
}

func schemaUpsertPayload(schema contract.Schema, databaseID string) map[string]any {
	return map[string]any{
		"databaseId": databaseID,
		"entities":   toEntities(schema),
	}
}

func toEntities(s contract.Schema) []schemaEntity {
	entities := make([]schemaEntity, 0, len(s.Tables))
	for _, t := range s.Tables {
		ent := schemaEntity{Name: t.Name}
		if len(t.Indexes) > 0 {
			for _, idx := range t.Indexes {
				ent.Indexes = append(ent.Indexes, map[string]any{"name": idx.Name})
			}
		}
		if len(t.Triggers) > 0 {
			for _, trig := range t.Triggers {
				ent.Triggers = append(ent.Triggers, map[string]any{"name": trig})
			}
		}
		if t.Meta != nil {
			ent.Meta = t.Meta
		}
		for _, f := range t.Fields {
			attr := schemaAttribute{
				Name:       f.Name,
				Type:       f.Type,
				IsNullable: f.Nullable,
			}
			ent.Attributes = append(ent.Attributes, attr)
			if f.Primary {
				ent.Identifier = &schemaIdentifier{Name: f.Name, Generator: "None", Type: f.Type}
			}
		}
		if len(t.Resolvers) > 0 {
			for _, r := range t.Resolvers {
				ent.Resolvers = append(ent.Resolvers, map[string]any{"name": r})
			}
		}
		entities = append(entities, ent)
	}
	return entities
}

func schemaFromEntities(items []any) contract.Schema {
	var tables []contract.Table
	for _, raw := range items {
		obj, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name, _ := obj["name"].(string)
		table := contract.Table{Name: name}
		var idName string
		if ident, ok := obj["identifier"].(map[string]any); ok {
			if n, ok := ident["name"].(string); ok {
				idName = n
			}
		}
		if attrs, ok := obj["attributes"].([]any); ok {
			for _, a := range attrs {
				attrObj, ok := a.(map[string]any)
				if !ok {
					continue
				}
				field := contract.Field{
					Name: attrObj["name"].(string),
				}
				if t, ok := attrObj["type"].(string); ok {
					field.Type = t
				}
				if n, ok := attrObj["isNullable"].(bool); ok {
					field.Nullable = n
				}
				if field.Name == idName {
					field.Primary = true
				}
				table.Fields = append(table.Fields, field)
			}
		}
		if res, ok := obj["resolvers"].([]any); ok {
			for _, r := range res {
				switch rv := r.(type) {
				case string:
					table.Resolvers = append(table.Resolvers, rv)
				case map[string]any:
					if name, ok := rv["name"].(string); ok {
						table.Resolvers = append(table.Resolvers, name)
					}
				}
			}
		}
		tables = append(tables, table)
	}
	return contract.Schema{Tables: tables}
}

func schemaFromTablesArray(items []any) contract.Schema {
	var tables []contract.Table
	for _, raw := range items {
		obj, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		t := contract.Table{Name: stringValue(obj["name"])}
		if fields, ok := obj["fields"].([]any); ok {
			for _, f := range fields {
				fm, ok := f.(map[string]any)
				if !ok {
					continue
				}
				field := contract.Field{
					Name:     stringValue(fm["name"]),
					Type:     stringValue(fm["type"]),
					Nullable: boolValue(fm["nullable"]),
					Primary:  boolValue(fm["primaryKey"]),
					Unique:   boolValue(fm["unique"]),
				}
				t.Fields = append(t.Fields, field)
			}
		}
		if res, ok := obj["resolvers"].([]any); ok {
			for _, r := range res {
				switch rv := r.(type) {
				case string:
					t.Resolvers = append(t.Resolvers, rv)
				case map[string]any:
					if name, ok := rv["name"].(string); ok {
						t.Resolvers = append(t.Resolvers, name)
					}
				}
			}
		}
		tables = append(tables, t)
	}
	return contract.Schema{Tables: tables}
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func boolValue(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func stripEntityText(v any) any {
	switch val := v.(type) {
	case map[string]any:
		delete(val, "entityText")
		for k, nested := range val {
			val[k] = stripEntityText(nested)
		}
		return val
	case []any:
		for i, nested := range val {
			val[i] = stripEntityText(nested)
		}
		return val
	default:
		return v
	}
}
