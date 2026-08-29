package contract

import (
	"encoding/json"
	"fmt"
)

// PageResult represents a single page of results along with the cursor for the next page.
type PageResult struct {
	Items      QueryResults `json:"items"`
	NextCursor string       `json:"nextCursor,omitempty"`
}

// UnmarshalJSON accepts both {items,nextCursor} (legacy) and {records,nextPage}
// shapes returned by the service.
func (p *PageResult) UnmarshalJSON(data []byte) error {
	type alias PageResult
	if err := json.Unmarshal(data, (*alias)(p)); err == nil && p.Items != nil {
		return nil
	}

	var alt struct {
		Records    QueryResults `json:"records"`
		NextPage   string       `json:"nextPage,omitempty"`
		NextCursor string       `json:"nextCursor,omitempty"`
	}
	if err := json.Unmarshal(data, &alt); err != nil {
		return err
	}
	p.Items = alt.Records
	if alt.NextPage != "" {
		p.NextCursor = alt.NextPage
	} else {
		p.NextCursor = alt.NextCursor
	}
	return nil
}

// UnmarshalMessagePackValue decodes either supported page envelope directly
// from the validated MessagePack value tree, preserving recursive int64 values
// in entity records.
func (p *PageResult) UnmarshalMessagePackValue(value any) error {
	wrapper, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("msgpack: cannot decode page result from %T", value)
	}

	if rawItems, exists := wrapper["items"]; exists {
		var items QueryResults
		if err := items.UnmarshalMessagePackValue(rawItems); err != nil {
			return err
		}
		// Match UnmarshalJSON: a non-null legacy items field selects the
		// {items,nextCursor} shape, including an empty array.
		if items != nil {
			nextCursor, err := messagePackCursor(wrapper, "nextCursor")
			if err != nil {
				return err
			}
			p.Items = items
			p.NextCursor = nextCursor
			return nil
		}
	}

	var records QueryResults
	if err := records.UnmarshalMessagePackValue(wrapper["records"]); err != nil {
		return err
	}
	nextPage, err := messagePackCursor(wrapper, "nextPage")
	if err != nil {
		return err
	}
	nextCursor, err := messagePackCursor(wrapper, "nextCursor")
	if err != nil {
		return err
	}
	p.Items = records
	if nextPage != "" {
		p.NextCursor = nextPage
	} else {
		p.NextCursor = nextCursor
	}
	return nil
}

func messagePackCursor(wrapper map[string]any, key string) (string, error) {
	value, exists := wrapper[key]
	if !exists || value == nil {
		return "", nil
	}
	cursor, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("msgpack: page field %q is %T; expected a string", key, value)
	}
	return cursor, nil
}
