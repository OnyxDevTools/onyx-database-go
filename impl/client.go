package impl

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
	"sync"
)

// Config mirrors contract.Config to avoid import churn.
type Config = contract.Config

type client struct {
	cfg        resolver.ResolvedConfig
	httpClient *httpclient.Client
	aiClient   *httpclient.Client
	now        func() time.Time
	sleep      func(time.Duration)
}

var (
	httpClientCache = struct {
		sync.Map
	}{}
	testHookAfterCacheLoad func()
)

func (c *client) tablePath(table string) string {
	return "/data/" + tableEscape(c.cfg.DatabaseID) + "/" + tableEscape(table)
}

func tableEscape(s string) string {
	return url.PathEscape(s)
}

func redactSecret(secret string) string {
	if len(secret) <= 4 {
		return "****"
	}
	return fmt.Sprintf("%s****", secret[:4])
}

func getCachedHTTPClient(baseURL string, baseHTTP *http.Client, logRequests, logResponses bool, signer httpclient.Signer, logger *log.Logger) *httpclient.Client {
	key := httpClientCacheKey(baseURL, baseHTTP, logRequests, logResponses, signer)
	if cached, ok := httpClientCache.Load(key); ok {
		return cached.(*httpclient.Client)
	}

	if testHookAfterCacheLoad != nil {
		testHookAfterCacheLoad()
	}

	hc := httpclient.New(baseURL, baseHTTP, httpclient.Options{
		Logger:       logger,
		LogRequests:  logRequests,
		LogResponses: logResponses,
		Signer:       signer,
	})

	if existing, loaded := httpClientCache.LoadOrStore(key, hc); loaded {
		return existing.(*httpclient.Client)
	}
	return hc
}

func httpClientCacheKey(baseURL string, baseHTTP *http.Client, logRequests, logResponses bool, signer httpclient.Signer) string {
	var basePtr string
	if baseHTTP != nil {
		basePtr = fmt.Sprintf("%p", baseHTTP)
	}
	return strings.Join([]string{
		baseURL,
		basePtr,
		fmt.Sprintf("%t", logRequests),
		fmt.Sprintf("%t", logResponses),
		signer.APIKey,
		signer.APISecret,
	}, "|")
}

func clearHTTPClientCache() {
	httpClientCache.Range(func(k, v any) bool {
		httpClientCache.Delete(k)
		return true
	})
}

// Init constructs a client using the provided configuration.
func Init(ctx context.Context, cfg Config) (contract.Client, error) {
	resolved, meta, err := resolver.Resolve(ctx, resolver.Config{
		DatabaseID:      cfg.DatabaseID,
		DatabaseBaseURL: cfg.DatabaseBaseURL,
		APIKey:          cfg.APIKey,
		APISecret:       cfg.APISecret,
		AIBaseURL:       cfg.AIBaseURL,
		CacheTTL:        cfg.CacheTTL,
		ConfigPath:      cfg.ConfigPath,
		LogRequests:     cfg.LogRequests,
		LogResponses:    cfg.LogResponses,
		Partition:       cfg.Partition,
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
	if os.Getenv("ONYX_DEBUG") == "true" {
		logger.Printf(
			"[onyx] init config: databaseId=%s baseUrl=%s aiBaseUrl=%s apiKey=%s apiSecret=%s partition=%q cacheTTL=%s logRequests=%t logResponses=%t configFile=%s",
			resolved.DatabaseID,
			resolved.DatabaseBaseURL,
			resolved.AIBaseURL,
			resolved.APIKey,
			redactSecret(resolved.APISecret),
			resolved.Partition,
			resolved.CacheTTL,
			logRequests,
			logResponses,
			meta.FilePath,
		)
	}
	signer := httpclient.Signer{
		APIKey:    resolved.APIKey,
		APISecret: resolved.APISecret,
	}

	hc := getCachedHTTPClient(resolved.DatabaseBaseURL, cfg.HTTPClient, logRequests, logResponses, signer, logger)
	ai := getCachedHTTPClient(resolved.AIBaseURL, cfg.HTTPClient, logRequests, logResponses, signer, logger)

	nowFn := time.Now
	if cfg.Clock != nil {
		nowFn = cfg.Clock
	}

	c := &client{cfg: resolved, httpClient: hc, aiClient: ai, now: nowFn}
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
	clearHTTPClientCache()
}

func (c *client) From(table string) contract.Query {
	return newQuery(c, table)
}

func (c *client) Search(queryText string, minScore ...float64) contract.Query {
	return newQuery(c, "ALL").Search(queryText, minScore...)
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
	if strings.TrimSpace(c.cfg.Partition) != "" {
		params := url.Values{}
		params.Set("partition", strings.TrimSpace(c.cfg.Partition))
		path += "?" + params.Encode()
	}
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
