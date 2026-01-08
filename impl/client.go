package impl

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

// Config mirrors contract.Config to avoid import churn.
type Config = contract.Config

type client struct {
	cfg        resolver.ResolvedConfig
	httpClient *httpclient.Client
	now        func() time.Time
	sleep      func(time.Duration)
}

func (c *client) tablePath(table string) string {
	return "/data/" + tableEscape(c.cfg.DatabaseID) + "/" + tableEscape(table)
}

func tableEscape(s string) string {
	return url.PathEscape(s)
}

// Init constructs a client using the provided configuration.
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

	// Match the TS SDK logging output (no timestamps/prefixes).
	logger := log.New(os.Stdout, "", 0)
	signer := httpclient.Signer{
		APIKey:    resolved.APIKey,
		APISecret: resolved.APISecret,
	}

	hc := httpclient.New(resolved.DatabaseBaseURL, cfg.HTTPClient, httpclient.Options{
		Logger:       logger,
		LogRequests:  logRequests,
		LogResponses: logResponses,
		Signer:       signer,
	})

	nowFn := time.Now
	if cfg.Clock != nil {
		nowFn = cfg.Clock
	}

	c := &client{cfg: resolved, httpClient: hc, now: nowFn}
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

func (c *client) Save(ctx context.Context, table string, entity any, relationships []string) (map[string]any, error) {
	path := c.tablePath(table)
	if len(relationships) > 0 {
		params := url.Values{}
		params.Set("relationships", strings.Join(relationships, ","))
		path += "?" + params.Encode()
	}
	var resp map[string]any
	if err := c.httpClient.DoJSON(ctx, http.MethodPut, path, entity, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) Delete(ctx context.Context, table, id string) error {
	path := c.tablePath(table) + "/" + tableEscape(id)
	return c.httpClient.DoJSON(ctx, http.MethodDelete, path, nil, nil)
}

func (c *client) BatchSave(ctx context.Context, table string, entities []any, batchSize int) error {
	return batchSave(ctx, c, table, entities, batchSize)
}

func (c *client) Schema(ctx context.Context) (contract.Schema, error) {
	return c.GetSchema(ctx, nil)
}

func (c *client) GetSchema(ctx context.Context, tables []string) (contract.Schema, error) {
	return fetchSchema(ctx, c, tables)
}

func (c *client) PublishSchema(ctx context.Context, schema contract.Schema) error {
	return publishSchema(ctx, c, schema, false)
}

func (c *client) UpdateSchema(ctx context.Context, schema contract.Schema, publish bool) error {
	return publishSchema(ctx, c, schema, publish)
}

func (c *client) ValidateSchema(ctx context.Context, schema contract.Schema) error {
	normalized := contract.NormalizeSchema(schema)
	path := "/schemas/" + tableEscape(c.cfg.DatabaseID) + "/validate"
	return c.httpClient.DoJSON(ctx, http.MethodPost, path, schemaUpsertPayload(normalized, c.cfg.DatabaseID), nil)
}

func (c *client) GetSchemaHistory(ctx context.Context) ([]contract.Schema, error) {
	var history []contract.Schema
	path := "/schemas/" + tableEscape(c.cfg.DatabaseID) + "/history"
	if err := c.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &history); err != nil {
		return nil, err
	}
	return history, nil
}
