package contract

import (
	"net/http"
	"time"
)

const (
	// WireFormatJSON keeps the existing JSON entity transport and is the default.
	WireFormatJSON = "json"
	// WireFormatMessagePack opts entity CRUD and query routes into MessagePack.
	WireFormatMessagePack = "msgpack"
)

// Config controls initialization of the SDK client.
type Config struct {
	DatabaseID      string
	DatabaseBaseURL string
	APIKey          string
	APISecret       string
	AIBaseURL       string
	CacheTTL        time.Duration
	ConfigPath      string
	LogRequests     bool
	LogResponses    bool
	Partition       string
	// WireFormat controls entity CRUD and query payloads. Empty is equivalent
	// to WireFormatJSON. Documents, schemas, secrets, and AI always use JSON.
	WireFormat string
	HTTPClient *http.Client
	Clock      func() time.Time
	Sleep      func(time.Duration)
}
