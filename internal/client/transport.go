package client

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/glwbr/brisa/pkg/logger"
)

// headersTransport adds default headers to outgoing requests.
type headersTransport struct {
	Next    http.RoundTripper
	Headers map[string]string
}

// RoundTrip implements the http.RoundTripper interface.
// It adds the configured headers to the request before delegating to the next transport.
func (t *headersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Apply default headers
	for k, v := range t.Headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}

	return t.next().RoundTrip(req)
}

// next returns the next RoundTripper, or http.DefaultTransport if nil.
func (t *headersTransport) next() http.RoundTripper {
	if t.Next != nil {
		return t.Next
	}
	return http.DefaultTransport
}

// loggingTransport logs HTTP request and response details.
// Logging is conditional based on the Debug flag.
type loggingTransport struct {
	Next   http.RoundTripper
	Logger logger.Logger
	Debug  bool
}

// RoundTrip implements the http.RoundTripper interface.
// It logs the request and response if debugging is enabled.
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.Debug {
		return t.next().RoundTrip(req)
	}

	start := time.Now()

	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	resp, err := t.next().RoundTrip(req)

	t.logRequest(req, reqBody, start)

	if err == nil && resp != nil {
		t.logResponse(resp)
	}

	return resp, err
}

// next returns the next RoundTripper, or http.DefaultTransport if nil.
func (t *loggingTransport) next() http.RoundTripper {
	if t.Next != nil {
		return t.Next
	}
	return http.DefaultTransport
}

// logRequest logs the HTTP request details using the configured logger.
func (t *loggingTransport) logRequest(req *http.Request, body []byte, start time.Time) {
	dump, _ := httputil.DumpRequestOut(req, false)

	fields := map[string]any{
		"method":   req.Method,
		"url":      req.URL.String(),
		"headers":  string(dump),
		"duration": time.Since(start).String(),
	}

	if len(body) > 0 {
		fields["body"] = string(body)
	}

	t.Logger.WithFields(fields).Debug("HTTP Request")
}

// logResponse logs the HTTP response details using the configured logger.
func (t *loggingTransport) logResponse(resp *http.Response) {
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

	if len(body) > 0 {
		fields["body"] = string(body)
	}

	t.Logger.WithFields(fields).Debug("HTTP Response")
}

// buildTransport constructs an HTTP transport chain based on the provided client configuration.
// It wraps http.DefaultTransport with optional layers such as header injection and request/response logging.
//
// Note: This implementation could be extended using a middleware-style pattern to enable
// dynamic composition of transport behaviors, while also decoupling it from ClientConfig.
// This would make it easier to plug in reusable layers for retries, tracing, metrics, etc...
func buildTransport(cfg *ClientConfig) http.RoundTripper {
	tr := http.DefaultTransport

	// WARN: Apply logging as the outermost wrapper
	tr = &loggingTransport{
		Next:   tr,
		Logger: cfg.Logger,
		Debug:  cfg.Debug,
	}

	tr = &headersTransport{
		Next:    tr,
		Headers: cfg.Headers,
	}

	return tr
}
