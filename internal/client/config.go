package client

import (
	"fmt"
	"maps"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/glwbr/brisa/pkg/logger"
)

// ClientConfig defines the configuration options for the HTTP client.
// All fields are optional, with sensible defaults provided by buildConfig.
type ClientConfig struct {
	BaseURL       *url.URL
	Timeout       time.Duration
	RetryAttempts int

	Headers map[string]string
	Jar     *cookiejar.Jar

	Logger logger.Logger
	Debug  bool

	CustomDoer Doer
}

// ClientOption defines a function that modifies the Config object.
// It is used to apply flexible and composable configuration settings.
type ClientOption func(*ClientConfig)

// WithTimeout sets the maximum duration for HTTP requests.
// The timeout includes connection time, any redirects, and reading the response body.
// A timeout <= 0 will be ignored and the default timeout will be used.
func WithTimeout(d time.Duration) ClientOption {
	return func(cfg *ClientConfig) {
		if d > 0 {
			cfg.Timeout = d
		}
	}
}

// WithBaseURL sets and normalizes the base URL for the client.
//
// This function ensures the provided baseURL is a valid absolute URL (with scheme and host).
// It removes any trailing slashes from the path to maintain consistency during URL resolution.
// If the provided baseURL is invalid or not absolute, an error is logged, and the URL is not set.
//
// Example:
//
//	"http://example.com/api/v1/" -> "http://example.com/api/v1"
//
// If the baseURL is empty or invalid, the client will use the default behavior (no base URL)
func WithBaseURL(baseURL string) ClientOption {
	return func(cfg *ClientConfig) {
		if baseURL == "" {
			return
		}

		u, err := normalizeBaseURL(baseURL)
		if err != nil {
			cfg.Logger.Error("invalid baseURL", "url", baseURL, "error", err)
			return
		}

		cfg.BaseURL = u
	}
}

// WithRetryAttempts configures the number of retries for failed requests.
// A value of 0 disables retries entirely. Negative values are ignored.
// Each retry follows an exponential backoff strategy.
func WithRetryAttempts(attempts int) ClientOption {
	return func(cfg *ClientConfig) {
		if attempts >= 0 {
			cfg.RetryAttempts = attempts
		}
	}
}

// WithHeaders sets default headers that will be included with every request.
// Existing headers with the same keys will be overwritten.
// The headers map is copied, so subsequent changes to the original won't affect the client.
func WithHeaders(headers map[string]string) ClientOption {
	return func(cfg *ClientConfig) {
		if len(headers) == 0 {
			return
		}
		if cfg.Headers == nil {
			cfg.Headers = make(map[string]string, len(headers))
		}
		maps.Copy(cfg.Headers, headers)
	}
}

// WithCookieJar provides a custom cookie jar for session management.
// If nil is provided or the jar is not set, cookies will not be persisted between requests.
func WithCookieJar(jar *cookiejar.Jar) ClientOption {
	return func(cfg *ClientConfig) {
		if jar != nil {
			cfg.Jar = jar
		}
	}
}

// WithLogger sets a custom logger for client operations.
// If nil is provided, the client will use a no-op logger by default.
func WithLogger(l logger.Logger) ClientOption {
	return func(cfg *ClientConfig) {
		if l != nil {
			cfg.Logger = l
		}
	}
}

// WithDebug enables verbose logging of HTTP requests and responses.
// When enabled, the logger will output detailed information including:
// - Full request/response headers
// - Request/response bodies (unless they contain binary data)
// - Timing information
func WithDebug(enable bool) ClientOption {
	return func(cfg *ClientConfig) { cfg.Debug = enable }
}

// WithCustomDoer allows injection of a custom HTTP client implementation.
// This can be used to mock the client for testing or provide special transport logic.
// The Doer interface must not be nil to take effect.
func WithCustomDoer(d Doer) ClientOption {
	return func(cfg *ClientConfig) {
		if d != nil {
			cfg.CustomDoer = d
		}
	}
}

// normalizeBaseURL parses and validates the given baseURL string.
// It ensures the URL is absolute (has scheme and host) and removes any trailing slash from the path.
// Returns a normalized *url.URL or an error if the input is invalid.
func normalizeBaseURL(baseURL string) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	if !u.IsAbs() {
		return nil, fmt.Errorf("base URL must be absolute (have scheme and host)")
	}

	u.Path = strings.TrimRight(u.Path, "/")
	return u, nil
}

// buildConfig constructs a Config with defaults and applies all provided options.
// Default values:
// - Timeout: defaultTimeout (package-level constant)
// - Logger: logger.NoOp{}
// - RetryAttempts: 3
// - Headers: Includes default User-Agent
// Any invalid option values will fall back to their defaults.
func buildConfig(opts ...ClientOption) *ClientConfig {
	cfg := &ClientConfig{
		Timeout:       defaultTimeout,
		Logger:        logger.NoOp{},
		RetryAttempts: 3,
		Headers:       map[string]string{"User-Agent": defaultUserAgent},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
