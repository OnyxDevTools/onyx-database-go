package contract

import "encoding/json"

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
