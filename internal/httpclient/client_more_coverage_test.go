package httpclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"testing"
)

func TestDoNewRequestError(t *testing.T) {
	orig := newRequestWithContext
	defer func() { newRequestWithContext = orig }()
	expected := errors.New("boom new request")
	newRequestWithContext = func(context.Context, string, string, io.Reader) (*http.Request, error) {
		return nil, expected
	}

	c := New("http://invalid", &http.Client{}, Options{})
	if err := c.DoJSON(context.Background(), http.MethodGet, "/x", nil, nil); !errors.Is(err, expected) {
		t.Fatalf("expected new request error, got %v", err)
	}
}

func TestDoSignError(t *testing.T) {
	orig := signRequest
	defer func() { signRequest = orig }()
	expected := errors.New("sign boom")
	signRequest = func(Signer, *http.Request, []byte) error { return expected }

	c := New("http://example.com", &http.Client{}, Options{})
	if err := c.DoJSON(context.Background(), http.MethodGet, "/x", nil, nil); !errors.Is(err, expected) {
		t.Fatalf("expected sign error, got %v", err)
	}
}

type errorBody struct{}

func (errorBody) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (errorBody) Close() error             { return nil }

type errorRoundTripper struct{}

func (errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       errorBody{},
	}, nil
}

func TestDoReadAllError(t *testing.T) {
	httpClient := &http.Client{Transport: errorRoundTripper{}}
	c := New("http://example.com", httpClient, Options{})
	if err := c.DoJSON(context.Background(), http.MethodGet, "/x", nil, nil); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected read error, got %v", err)
	}
}

type okRoundTripper struct{}

func (okRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
	}, nil
}

func TestDoLogsResponseBody(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", 0)
	httpClient := &http.Client{Transport: okRoundTripper{}}
	c := New("http://example.com", httpClient, Options{Logger: logger, LogResponses: true})
	if err := c.DoJSON(context.Background(), http.MethodGet, "/x", nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`{"ok":true}`)) {
		t.Fatalf("expected response body logged, got %s", buf.String())
	}
}
