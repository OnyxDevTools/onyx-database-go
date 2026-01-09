package contract

import (
	"encoding/json"
	"sort"
)

func normalizeResolvers(res []Resolver) []Resolver {
	out := make([]Resolver, len(res))
	copy(out, res)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ParseSchemaJSON parses a schema document from JSON bytes.
func ParseSchemaJSON(data []byte) (Schema, error) {
	var s Schema
	if err := json.Unmarshal(data, &s); err == nil && len(s.Tables) > 0 {
		return s, nil
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return Schema{}, err
	}

	if nested, ok := raw["schema"].(map[string]any); ok {
		raw = nested
	}

	if tables, ok := raw["tables"].([]any); ok {
		return schemaFromTablesArray(tables), nil
	}

	if entities, ok := raw["entities"].([]any); ok {
		return schemaFromEntities(entities), nil
	}

	return s, nil
}

// NormalizeSchema returns a copy of the schema with deterministic ordering.
func NormalizeSchema(s Schema) Schema {
	normalized := Schema{Tables: make([]Table, len(s.Tables))}
	copy(normalized.Tables, s.Tables)

	sort.Slice(normalized.Tables, func(i, j int) bool {
		return normalized.Tables[i].Name < normalized.Tables[j].Name
	})

	for i := range normalized.Tables {
		fields := make([]Field, len(normalized.Tables[i].Fields))
		copy(fields, normalized.Tables[i].Fields)
		sort.Slice(fields, func(a, b int) bool {
			return fields[a].Name < fields[b].Name
		})
		normalized.Tables[i].Fields = fields

		if len(normalized.Tables[i].Resolvers) > 0 {
			normalized.Tables[i].Resolvers = normalizeResolvers(normalized.Tables[i].Resolvers)
		}
		if len(normalized.Tables[i].Indexes) > 0 {
			indexes := make([]Index, len(normalized.Tables[i].Indexes))
			copy(indexes, normalized.Tables[i].Indexes)
			sort.Slice(indexes, func(a, b int) bool { return indexes[a].Name < indexes[b].Name })
			normalized.Tables[i].Indexes = indexes
		}
		if len(normalized.Tables[i].Triggers) > 0 {
			trigs := make([]string, len(normalized.Tables[i].Triggers))
			copy(trigs, normalized.Tables[i].Triggers)
			sort.Strings(trigs)
			normalized.Tables[i].Triggers = trigs
		}
	}

	return normalized
}

func schemaFromEntities(items []any) Schema {
	var tables []Table
	for _, raw := range items {
		obj, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name, _ := obj["name"].(string)
		table := Table{Name: name}
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
				field := Field{
					Name: stringValue(attrObj["name"]),
					Type: stringValue(attrObj["type"]),
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
				if rm, ok := r.(map[string]any); ok {
					if name, ok := rm["name"].(string); ok {
						table.Resolvers = append(table.Resolvers, Resolver{
							Name:     name,
							Resolver: stringValue(rm["resolver"]),
							Meta:     mapValue(rm["meta"]),
						})
					}
				} else if name, ok := r.(string); ok {
					table.Resolvers = append(table.Resolvers, Resolver{Name: name})
				}
			}
		}
		if idxs, ok := obj["indexes"].([]any); ok {
			for _, idx := range idxs {
				if im, ok := idx.(map[string]any); ok {
					if name, ok := im["name"].(string); ok {
						table.Indexes = append(table.Indexes, Index{Name: name})
					}
				}
			}
		}
		if trigs, ok := obj["triggers"].([]any); ok {
			for _, t := range trigs {
				if name, ok := t.(string); ok {
					table.Triggers = append(table.Triggers, name)
				} else if tm, ok := t.(map[string]any); ok {
					if name, ok := tm["name"].(string); ok {
						table.Triggers = append(table.Triggers, name)
					}
				}
			}
		}
		if meta, ok := obj["meta"].(map[string]any); ok {
			table.Meta = meta
		}
		tables = append(tables, table)
	}
	return Schema{Tables: tables}
}

func schemaFromTablesArray(items []any) Schema {
	var tables []Table
	for _, raw := range items {
		obj, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		t := Table{Name: stringValue(obj["name"])}
		if fields, ok := obj["fields"].([]any); ok {
			for _, f := range fields {
				fm, ok := f.(map[string]any)
				if !ok {
					continue
				}
				field := Field{
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
				switch v := r.(type) {
				case string:
					t.Resolvers = append(t.Resolvers, Resolver{Name: v})
				case map[string]any:
					if name, ok := v["name"].(string); ok {
						t.Resolvers = append(t.Resolvers, Resolver{
							Name:     name,
							Resolver: stringValue(v["resolver"]),
							Meta:     mapValue(v["meta"]),
						})
					}
				}
			}
		}
		if idxs, ok := obj["indexes"].([]any); ok {
			for _, idx := range idxs {
				if im, ok := idx.(map[string]any); ok {
					if name, ok := im["name"].(string); ok {
						t.Indexes = append(t.Indexes, Index{Name: name})
					}
				}
			}
		}
		if trigs, ok := obj["triggers"].([]any); ok {
			for _, trg := range trigs {
				if tm, ok := trg.(map[string]any); ok {
					if name, ok := tm["name"].(string); ok {
						t.Triggers = append(t.Triggers, name)
					}
				} else if name, ok := trg.(string); ok {
					t.Triggers = append(t.Triggers, name)
				}
			}
		}
		if meta, ok := obj["meta"].(map[string]any); ok {
			t.Meta = meta
		}
		tables = append(tables, t)
	}
	return Schema{Tables: tables}
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

func mapValue(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}
