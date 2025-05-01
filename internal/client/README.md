# 🌐 HTTP Client

A simple, configurable HTTP client with comprehensive features for making API requests.
Used internally by **Brisa** to interact with SEFAZ's NFC-e endpoints, but flexible enough for other domains.
Built for learning and extensibility purposes, this package may be refactored in the future for simplicity.

P.S.: It only provides needed methods right now, `GET` and `POST` but other methods might be easily added.

## ✨ Features

- Debug logging with configurable logger
- Functional options for client and transport configuration
- Mockable `Doer` interface for testing
- Cookie and session management
- Context support for timeout and cancellation
- Sensible defaults with type-safe configuration
- Composable transport chain for request/response (middleware pattern) Foo -> Bar -> http.DefaultTransport

## 🚀 Getting Started

### Basic Request

```go
package bar

resp, err := client.Get("https://api.example.com/data")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()
```

### Custom Client Setup

```go
package foo

c, err := client.New(
    client.WithBaseURL("https://api.example.com/v1"),
    client.WithTimeout(30*time.Second),
    client.WithRetryAttempts(3),
    client.WithDebug(true),
)
```

### Per-request Configuration

```go
package bar

&client.RequestConfig{
    Params:  url.Values{"q": {"search"}},
    Body:    strings.NewReader(`{"data":1}`),
    Headers: map[string]string{"X-Custom": "value"},
}
```

## 🧪 Examples

### JSON POST

```go
package foo

body := strings.NewReader(`{"name":"Brisa"}`)
resp, err := c.Post(ctx, "/items", &client.RequestConfig{
    Body: body,
    Headers: map[string]string{
        "Content-Type": "application/json",
    },
})
```

### GET with Query Params

```go
package bar

params := url.Values{}
params.Set("limit", "5")
params.Set("status", "active")

resp, err := c.Get(ctx, "/records", &client.RequestConfig{Params: params})
```

## 🧠 Advanced Usage

### Custom Transport Chain

Handle cross-cutting concerns like logging and headers by composing transports:

```go
package foo

client, err := client.New(
    client.WithLogger(myLogger),
    client.WithDebug(true),
    client.WithHeaders(map[string]string{
        "User-Agent": "Brisa/1.0",
        "X-API-Key": "abc123",
    }),
)
```

Under the hood:

> I don't really know if I should expose this, but we'll figure it out.

```go
package bar

httpClient := &http.Client{
    Timeout: cfg.Timeout,
    Transport: NewTransport(
        WithHeadersTransport(cfg.Headers),
        WithLogging(cfg.Logger, cfg.Debug),
    ),
}
```

### Creating Custom Transports

```go
package client

import "net/http"

func WithRetry(maxRetries int, retryableCodes []int) TransportOption {
	return func(next http.RoundTripper) http.RoundTripper {
		return &RetryTransport{
			Next:                 next,
			MaxRetries:           maxRetries,
			RetryableStatusCodes: retryableCodes,
		}
	}
}

// Example transport
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Implement retry logic...
}
```

## 📊 Logging & Error Handling

### Logging

Enable verbose output by enabling debug mode and setting a logger:

```go
package foo

client.New(
    client.WithDebug(true),
    client.WithLogger(myLogger),
)
```

Logs include:

- Request method, URL, and headers
- Request and response bodies
- Status codes and latency

### Error Type Matching

```go
package foo

resp, err := c.Get(ctx, "/path", nil)
if err != nil {
    var httpErr *errors.HTTPError
    if errors.As(err, &httpErr) {
        log.Printf("HTTP %d: %s", httpErr.StatusCode(), httpErr.Message)
    } else {
        log.Printf("Request failed: %v", err)
    }
}
```

## 🐛 Testing

### Mocking with `Doer`

```go
package client

type mockDoer struct {
    response *http.Response
    err      error
}

func (m *mockDoer) Do(*http.Request) (*http.Response, error) {
    return m.response, m.err
}

client.New(
    client.WithCustomDoer(&mockDoer{
        response: &http.Response{
            StatusCode: 200,
            Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
        },
    }),
)
```

### Testing Custom Transports

```go
package client

import (
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestHeadersTransport(t *testing.T) {
	mockTransport := &mockTransport{}
	headersTransport := &HeadersTransport{
		Next:    mockTransport,
		Headers: map[string]string{"X-Test": "value"},
	}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	_, err := headersTransport.RoundTrip(req)

	assert.NoError(t, err)
	assert.Equal(t, "value", req.Header.Get("X-Test"))
}
```

## 💡 Best Practices

- Reuse clients: they're thread-safe.
- Use `context.Context` for timeouts and cancellation.
- Always `defer resp.Body.Close()`.
- Use `errors.As()` to match HTTP errors.
- Tune timeouts based on external service behavior.
- Compose single-responsibility transports.

## 🛠️ Defaults Summary

| Setting        | Default           |
| -------------- | ----------------- |
| Timeout        | 10s               |
| Retry Attempts | 3 (with backoff)  |
| Headers        | `User-Agent` only |
| Logger         | No-op             |
| Debug Mode     | Off               |

## 🔧 Internal Architecture

The client is built on a modular architecture where configuration is done using functional options:

```go
type ClientOption func(*ClientConfig)
type TransportOption func(http.RoundTripper) http.RoundTripper
// TODO: type RequestOption func(*RequestConfig)
```

This enables extensibility while keeping concerns separated.
