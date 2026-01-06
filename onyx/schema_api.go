package onyx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func fetchSchema(ctx context.Context, c *client) (contract.Schema, error) {
	var raw map[string]any
	if err := c.httpClient.DoJSON(ctx, http.MethodGet, "/schema", nil, &raw); err != nil {
		return contract.Schema{}, err
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
