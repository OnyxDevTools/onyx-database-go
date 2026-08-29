package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/internal/msgpack"
)

// Options controls the behavior of the HTTP client wrapper.
type Options struct {
	Logger       *log.Logger
	LogRequests  bool
	LogResponses bool
	Signer       Signer
}

// Client wraps an http.Client with helpers for Onyx API communication.
type Client struct {
	baseURL      string
	httpClient   *http.Client
	logger       *log.Logger
	logRequests  bool
	logResponses bool
	signer       Signer
}

var (
	newRequestWithContext = http.NewRequestWithContext
	signRequest           = func(s Signer, req *http.Request, body []byte) error {
		return s.Sign(req, body)
	}
)

// New constructs a Client.
func New(baseURL string, httpClient *http.Client, opts Options) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if opts.Logger == nil {
		opts.Logger = log.New(io.Discard, "", log.LstdFlags)
	}
	return &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		httpClient:   httpClient,
		logger:       opts.Logger,
		logRequests:  opts.LogRequests,
		logResponses: opts.LogResponses,
		signer:       opts.Signer,
	}
}

// Logger returns the configured logger (never nil).
func (c *Client) Logger() *log.Logger {
	return c.logger
}

// LogResponses reports whether response logging is enabled.
func (c *Client) LogResponses() bool {
	return c.logResponses
}

// DoJSON executes an HTTP request and decodes the JSON response.
func (c *Client) DoJSON(ctx context.Context, method, path string, reqBody any, respBody any) error {
	resp, body, err := c.do(ctx, method, path, reqBody, false)
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if respBody == nil || len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, respBody)
}

// DoEntity executes an entity request in the configured wire format. A server
// may answer a MessagePack request with JSON; the actual response Content-Type
// selects the decoder.
func (c *Client) DoEntity(ctx context.Context, method, path string, reqBody any, respBody any, wireFormat string) error {
	if wireFormat != contract.WireFormatMessagePack {
		return c.DoJSON(ctx, method, path, reqBody, respBody)
	}

	body, err := encodeMessagePackBody(reqBody)
	if err != nil {
		return err
	}
	resp, responseBody, err := c.doEncoded(
		ctx,
		method,
		path,
		body,
		msgpack.MediaType,
		msgpack.MediaType+", application/json;q=0.9",
		false,
		true,
	)
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		return err
	}
	if respBody == nil || len(responseBody) == 0 {
		return nil
	}
	if isMessagePackMediaType(resp.Header.Get("Content-Type")) {
		return msgpack.Unmarshal(responseBody, respBody)
	}
	return json.Unmarshal(responseBody, respBody)
}

// DoStream executes an HTTP request and returns the response for streaming consumption.
func (c *Client) DoStream(ctx context.Context, method, path string, reqBody any) (*http.Response, error) {
	resp, _, err := c.do(ctx, method, path, reqBody, true)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return nil, err
	}
	return resp, nil
}

// DoEntityStream starts an entity stream in the configured wire format.
func (c *Client) DoEntityStream(ctx context.Context, method, path string, reqBody any, wireFormat string) (*http.Response, error) {
	if wireFormat != contract.WireFormatMessagePack {
		return c.DoStream(ctx, method, path, reqBody)
	}
	body, err := encodeMessagePackBody(reqBody)
	if err != nil {
		return nil, err
	}
	resp, _, err := c.doEncoded(
		ctx,
		method,
		path,
		body,
		msgpack.MediaType,
		msgpack.MediaType+", application/json;q=0.9",
		true,
		true,
	)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return nil, err
	}
	return resp, nil
}

func (c *Client) do(ctx context.Context, method, path string, reqBody any, streaming bool) (*http.Response, []byte, error) {
	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return nil, nil, err
		}
	}
	accept := ""
	if streaming {
		accept = "text/event-stream"
	}
	return c.doEncoded(ctx, method, path, buf.Bytes(), "application/json", accept, streaming, false)
}

func (c *Client) doEncoded(
	ctx context.Context,
	method, path string,
	body []byte,
	contentType, accept string,
	streaming, binaryRequest bool,
) (*http.Response, []byte, error) {
	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/")

	req, err := newRequestWithContext(ctx, method, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	// Preserve the legacy JSON header behavior, but do not claim that a
	// bodyless MessagePack GET/DELETE carries an encoded representation.
	if contentType != "" && (len(body) > 0 || !binaryRequest) {
		req.Header.Set("Content-Type", contentType)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	if err := signRequest(c.signer, req, body); err != nil {
		return nil, nil, err
	}
	if binaryRequest && len(body) == 0 {
		// The signer supplies legacy JSON defaults; remove that default when
		// this binary entity request has no representation body at all.
		req.Header.Del("Content-Type")
	}

	if c.logRequests {
		c.logger.Printf("[onyx] %s %s", method, req.URL.String())
		if len(body) > 0 {
			if binaryRequest {
				c.logger.Printf("[onyx] <%d-byte MessagePack body>", len(body))
			} else {
				c.logger.Printf("[onyx] %s", strings.TrimSpace(string(body)))
			}
		}
		c.logger.Printf(
			"[onyx] Headers: {x-onyx-key: '%s', x-onyx-secret: '%s', Accept: '%s', Content-Type: '%s'}",
			req.Header.Get("x-onyx-key"),
			redactedSecret(req.Header.Get("x-onyx-secret")),
			req.Header.Get("Accept"),
			req.Header.Get("Content-Type"),
		)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if c.logResponses {
			c.logger.Printf("[onyx] %s", resp.Status)
			if len(data) > 0 {
				c.logger.Printf("[onyx] %s", strings.TrimSpace(string(data)))
			}
		}
		return resp, nil, parseError(req.Context(), resp.StatusCode, data)
	}

	if streaming {
		if c.logResponses {
			c.logger.Printf("[onyx] %s", resp.Status)
		}
		return resp, nil, nil
	}

	reader := io.Reader(resp.Body)
	if isMessagePackMediaType(resp.Header.Get("Content-Type")) {
		reader = io.LimitReader(resp.Body, msgpack.MaxMessageLength+1)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return resp, nil, err
	}
	if isMessagePackMediaType(resp.Header.Get("Content-Type")) && len(data) > msgpack.MaxMessageLength {
		return resp, nil, fmt.Errorf("msgpack: message exceeds %d-byte limit", msgpack.MaxMessageLength)
	}

	if c.logResponses {
		c.logger.Printf("[onyx] %s", resp.Status)
		if len(data) > 0 {
			if isMessagePackMediaType(resp.Header.Get("Content-Type")) {
				c.logger.Printf("[onyx] <%d-byte MessagePack body>", len(data))
			} else {
				c.logger.Printf("[onyx] %s", strings.TrimSpace(string(data)))
			}
		}
	}

	return resp, data, nil
}

func encodeMessagePackBody(reqBody any) ([]byte, error) {
	if reqBody == nil {
		return nil, nil
	}
	return msgpack.Marshal(reqBody)
}

// IsMessagePackResponse reports whether a response advertises the entity
// MessagePack media type. It is shared with the stream iterator.
func IsMessagePackResponse(resp *http.Response) bool {
	return resp != nil && isMessagePackMediaType(resp.Header.Get("Content-Type"))
}

func isMessagePackMediaType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		mediaType = strings.TrimSpace(strings.SplitN(value, ";", 2)[0])
	}
	return strings.EqualFold(mediaType, msgpack.MediaType)
}

// redactedSecret returns a redacted representation to avoid leaking credentials.
func redactedSecret(secret string) string {
	if len(secret) <= 4 {
		return "****"
	}
	return fmt.Sprintf("%s****", secret[:4])
}
