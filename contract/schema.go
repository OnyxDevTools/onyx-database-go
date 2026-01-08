package contract

// Field represents a single column in a table definition.
type Field struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable,omitempty"`
	Primary  bool   `json:"primaryKey,omitempty"`
	Unique   bool   `json:"unique,omitempty"`
}

// Table represents a database table with fields.
type Table struct {
	Name      string         `json:"name"`
	Fields    []Field        `json:"fields"`
	Resolvers []string       `json:"resolvers,omitempty"`
	Indexes   []Index        `json:"indexes,omitempty"`
	Triggers  []string       `json:"triggers,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// Field retrieves a field by name, if present.
func (t Table) Field(name string) (Field, bool) {
	for _, f := range t.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

// Schema describes the collection of tables available to a client.
type Schema struct {
	Tables []Table `json:"tables"`
}

// Table returns a table by name.
func (s Schema) Table(name string) (Table, bool) {
	for _, t := range s.Tables {
		if t.Name == name {
			return t, true
		}
	}
	return Table{}, false
}
