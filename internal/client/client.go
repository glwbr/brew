// Package client provides a flexible HTTP client for making API requests
// with configurable transport chains, logging, and request handling.
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

// Doer performs HTTP requests, allowing for custom implementations and testing mocks.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client is an HTTP client with support for base URLs, middleware chains, and logging.
type Client struct {
	doer    Doer
	baseURL *url.URL
	logger  logger.Logger
	config  *ClientConfig
}

// New creates a Client with the provided options.
// It uses sensible defaults that can be overridden with ClientOption functions.
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

// createDefaultDoer builds an http.Client with the configured options.
func createDefaultDoer(cfg *ClientConfig) Doer {
	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: buildTransport(cfg),
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
		log.Fatalf("failed to initialize default HTTP client: %v", err)
	}
}
