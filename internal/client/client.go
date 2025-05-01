// Package client provides a flexible HTTP client with common functionality
// for making API requests with configurable behavior.
package client

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/glwbr/brisa/pkg/logger"
)

const (
	defaultTimeout   = 10 * time.Second
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

var defaultClient *Client

// Doer is an interface for executing HTTP requests.
// It allows for easy mocking in tests and custom HTTP client implementations.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client is a generic HTTP client with common functionality for API requests.
type Client struct {
	doer    Doer
	baseURL *url.URL
	logger  logger.Logger
	config  *ClientConfig
}

// New creates an API HTTPClient with optional configuration and sensible defaults.
// Use the With* option functions to customize the client's behavior.
// when will this error? we must check to return it correctly
// it might fail when the user creates a CustomClient and doesnt
// set tls coinfig correctly, but how to get these errors ?
func New(opts ...ClientOption) (*Client, error) {
	var doer Doer

	cfg := buildConfig(opts...)
	if cfg.CustomDoer != nil {
		doer = cfg.CustomDoer
	} else {
		doer = createDefaultDoer(cfg)
	}

	return &Client{
		doer:    doer,
		baseURL: cfg.BaseURL,
		logger:  cfg.Logger,
		config:  cfg,
	}, nil
}

func createDefaultDoer(cfg *ClientConfig) Doer {
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: NewTransport(
			WithHeadersTransport(cfg.Headers),
			WithLogging(cfg.Logger, cfg.Debug),
		),
	}

	if cfg.Jar != nil {
		client.Jar = cfg.Jar
	}

	return client
}

func init() {
	var err error

	defaultClient, err = New()
	if err != nil {
		log.Printf("failed to initialize HTTP client: %v", err)
	}
}
