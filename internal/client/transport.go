package client

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/glwbr/brisa/pkg/logger"
)

// TransportOption defines a function that modifies a transport chain.
type TransportOption func(http.RoundTripper) http.RoundTripper

// WithHeadersTransport returns a transport that injects headers into requests.
func WithHeadersTransport(headers map[string]string) TransportOption {
	return func(next http.RoundTripper) http.RoundTripper {
		return &HeadersTransport{
			Next:    next,
			Headers: headers,
		}
	}
}

// WithLogging returns a transport that logs HTTP requests and responses.
func WithLogging(logger logger.Logger, debug bool) TransportOption {
	return func(next http.RoundTripper) http.RoundTripper {
		return &LoggingTransport{
			Next:   next,
			Logger: logger,
			Debug:  debug,
		}
	}
}

// NewTransport creates a new transport chain with the given options.
func NewTransport(opts ...TransportOption) http.RoundTripper {
	transport := http.DefaultTransport

	// Apply options in reverse to maintain the expected order
	// (last option becomes the outermost transport)
	for i := len(opts) - 1; i >= 0; i-- {
		transport = opts[i](transport)
	}

	return transport
}

// HeadersTransport injects headers into requests.
type HeadersTransport struct {
	Next    http.RoundTripper
	Headers map[string]string
}

func (t *HeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Apply default headers
	for k, v := range t.Headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}

	return t.next().RoundTrip(req)
}

func (t *HeadersTransport) next() http.RoundTripper {
	if t.Next != nil {
		return t.Next
	}
	return http.DefaultTransport
}

// LoggingTransport logs HTTP requests and responses.
type LoggingTransport struct {
	Next   http.RoundTripper
	Logger logger.Logger
	Debug  bool
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.Debug {
		return t.next().RoundTrip(req)
	}

	start := time.Now()

	// Preserve the request body for logging
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	// Execute the request
	resp, err := t.next().RoundTrip(req)

	// Log request even if there was an error
	t.logRequest(req, reqBody, start)

	// Only log response if we got one
	if err == nil && resp != nil {
		t.logResponse(resp)
	}

	return resp, err
}

func (t *LoggingTransport) next() http.RoundTripper {
	if t.Next != nil {
		return t.Next
	}
	return http.DefaultTransport
}

func (t *LoggingTransport) logRequest(req *http.Request, body []byte, start time.Time) {
	dump, _ := httputil.DumpRequestOut(req, false)

	fields := map[string]any{
		"method":   req.Method,
		"url":      req.URL.String(),
		"headers":  string(dump),
		"duration": time.Since(start).String(),
	}

	// Add body only if it exists
	if len(body) > 0 {
		fields["body"] = string(body)
	}

	t.Logger.WithFields(fields).Debug("HTTP Request")
}

func (t *LoggingTransport) logResponse(resp *http.Response) {
	dump, _ := httputil.DumpResponse(resp, false)

	var body []byte
	if resp.Body != nil {
		body, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	fields := map[string]any{
		"status":  resp.Status,
		"headers": string(dump),
	}

	// Add body only if it exists
	if len(body) > 0 {
		fields["body"] = string(body)
	}

	t.Logger.WithFields(fields).Debug("HTTP Response")
}
