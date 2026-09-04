package contract

import (
	"net/http"
	"time"
)

const (
	// WireFormatJSON opts entity CRUD and query routes into JSON.
	WireFormatJSON = "json"
	// WireFormatMessagePack uses the default MessagePack entity transport.
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
	// to WireFormatMessagePack. Documents, schemas, secrets, and AI always use JSON.
	WireFormat string
	HTTPClient *http.Client
	Clock      func() time.Time
	Sleep      func(time.Duration)
}
