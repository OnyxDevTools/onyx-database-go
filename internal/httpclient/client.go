package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

func (c *Client) do(ctx context.Context, method, path string, reqBody any, streaming bool) (*http.Response, []byte, error) {
	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/")

	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return nil, nil, err
		}
	}

	req, err := newRequestWithContext(ctx, method, fullURL, &buf)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if streaming {
		req.Header.Set("Accept", "text/event-stream")
	}

	if err := signRequest(c.signer, req, buf.Bytes()); err != nil {
		return nil, nil, err
	}

	if c.logRequests {
		c.logger.Printf("[onyx] %s %s", method, req.URL.String())
		if buf.Len() > 0 {
			c.logger.Printf("[onyx] %s", strings.TrimSpace(buf.String()))
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}

	if c.logResponses {
		c.logger.Printf("[onyx] %s", resp.Status)
		if len(data) > 0 {
			c.logger.Printf("[onyx] %s", strings.TrimSpace(string(data)))
		}
	}

	return resp, data, nil
}

// redactedSecret returns a redacted representation to avoid leaking credentials.
func redactedSecret(secret string) string {
	if len(secret) <= 4 {
		return "****"
	}
	return fmt.Sprintf("%s****", secret[:4])
}
