package onyx

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
	"github.com/OnyxDevTools/onyx-database-go/onyx/resolver"
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
	HTTPClient      *http.Client
	Clock           func() time.Time
	Sleep           func(time.Duration)
}

type client struct {
	cfg        resolver.ResolvedConfig
	httpClient *httpclient.Client
	now        func() time.Time
	sleep      func(time.Duration)
}

// Init constructs a contract.Client using the provided configuration.
func Init(ctx context.Context, cfg Config) (contract.Client, error) {
	resolved, _, err := resolver.Resolve(ctx, resolver.Config{
		DatabaseID:      cfg.DatabaseID,
		DatabaseBaseURL: cfg.DatabaseBaseURL,
		APIKey:          cfg.APIKey,
		APISecret:       cfg.APISecret,
		CacheTTL:        cfg.CacheTTL,
		ConfigPath:      cfg.ConfigPath,
		LogRequests:     cfg.LogRequests,
		LogResponses:    cfg.LogResponses,
	})
	if err != nil {
		return nil, err
	}

	logRequests := resolved.LogRequests
	logResponses := resolved.LogResponses
	if os.Getenv("ONYX_DEBUG") == "true" {
		logRequests = true
		logResponses = true
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	signer := httpclient.Signer{
		APIKey:    resolved.APIKey,
		APISecret: resolved.APISecret,
		Now: func() time.Time {
			if cfg.Clock != nil {
				return cfg.Clock()
			}
			return time.Now()
		},
		RequestID: func() string { return uuid.NewString() },
	}

	hc := httpclient.New(resolved.DatabaseBaseURL, cfg.HTTPClient, httpclient.Options{
		Logger:       logger,
		LogRequests:  logRequests,
		LogResponses: logResponses,
		Signer:       signer,
	})

	c := &client{cfg: resolved, httpClient: hc, now: signer.Now}
	if cfg.Sleep != nil {
		c.sleep = cfg.Sleep
	} else {
		c.sleep = time.Sleep
	}

	return c, nil
}

// InitWithDatabaseID initializes a client using only the database ID and environment/file configuration.
func InitWithDatabaseID(ctx context.Context, databaseID string) (contract.Client, error) {
	return Init(ctx, Config{DatabaseID: databaseID})
}

// ClearConfigCache clears the resolver cache.
func ClearConfigCache() {
	resolver.ClearCache()
}

func (c *client) From(table string) contract.Query {
	return newQuery(c, table)
}

func (c *client) Cascade(spec contract.CascadeSpec) contract.CascadeClient {
	return &cascadeClient{client: c, spec: spec}
}

func (c *client) Save(ctx context.Context, table string, entity any) error {
	path := "/tables/" + table + "/save"
	return c.httpClient.DoJSON(ctx, http.MethodPost, path, entity, nil)
}

func (c *client) Delete(ctx context.Context, table, id string) error {
	path := "/tables/" + table + "/" + id
	return c.httpClient.DoJSON(ctx, http.MethodDelete, path, nil, nil)
}

func (c *client) BatchSave(ctx context.Context, table string, entities []any, batchSize int) error {
	return batchSave(ctx, c, table, entities, batchSize)
}

func (c *client) Schema(ctx context.Context) (contract.Schema, error) {
	return fetchSchema(ctx, c)
}

func (c *client) PublishSchema(ctx context.Context, schema contract.Schema) error {
	normalized := contract.NormalizeSchema(schema)
	return c.httpClient.DoJSON(ctx, http.MethodPost, "/schema", normalized, nil)
}
