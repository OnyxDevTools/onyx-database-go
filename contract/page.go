package contract

// PageResult represents a single page of results along with the cursor for the next page.
type PageResult struct {
	Items      QueryResults `json:"items"`
	NextCursor string       `json:"nextCursor,omitempty"`
}
