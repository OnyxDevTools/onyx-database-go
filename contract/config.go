package contract

import (
	"net/http"
	"time"
)

// Config controls initialization of the SDK client.
type Config struct {
	DatabaseID      string
	DatabaseBaseURL string
	APIKey          string
	APISecret       string
	CacheTTL        time.Duration
	ConfigPath      string
	LogRequests     bool
	LogResponses    bool
	Partition       string
	HTTPClient      *http.Client
	Clock           func() time.Time
	Sleep           func(time.Duration)
}
