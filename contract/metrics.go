package contract

// MetricsParams defines the identifiers that can be used to fetch metrics.
type MetricsParams struct {
	// UID is a unique user identifier.
	UID string `json:"uid,omitempty"`
	// Email is a user email address.
	Email string `json:"email,omitempty"`
}

// Metrics represents a flexible metrics payload returned by the API.
type Metrics map[string]any
