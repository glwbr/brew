package client

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/glwbr/brisa/pkg/errors"
)

// RequestConfig provides configuration for individual requests.
type RequestConfig struct {
	Params  url.Values
	Body    io.Reader
	Headers map[string]string
}

// Get performs an HTTP GET request to the specified path or URL.
// If a relative path is provided, it will be resolved against the base URL if set.
func (c *Client) Get(ctx context.Context, path string, opts *RequestConfig) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, opts)
}

// Get is a convenience function for performing a simple HTTP GET request using the default client.
// It uses a background context and does not support additional configuration.
func Get(url string) (*http.Response, error) {
	if defaultClient == nil {
		return nil, errors.New("default client not initialized")
	}

	ctx := context.Background()
	return defaultClient.Get(ctx, url, nil)
}

// Post performs an HTTP POST request to the specified path or URL.
// The request body should be provided in the options parameter.
func (c *Client) Post(ctx context.Context, path string, opts *RequestConfig) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, path, opts)
}

// do is the core method for executing HTTP requests with the configured client.
func (c *Client) do(ctx context.Context, method, urlOrPath string, opts *RequestConfig) (*http.Response, error) {
	if opts == nil {
		opts = &RequestConfig{}
	}

	u, err := c.resolveURL(urlOrPath, opts.Params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve URL")
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), opts.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	// Apply request-specific headers
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// Perform the request
	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, errors.NewHTTPError(nil, err, "request failed")
	}

	// Check if the response indicates an error
	if resp.StatusCode >= 400 {
		return resp, errors.NewHTTPError(resp, nil, "request returned error status")
	}

	return resp, err
}

// resolveURL handles path or full URL resolution against the base URL.
func (c *Client) resolveURL(pathOrURL string, queryParams url.Values) (*url.URL, error) {
	// Parse the input (could be a path or full URL)
	u, err := url.Parse(pathOrURL)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid URL or path: %s", pathOrURL)
	}

	// If it's already a complete URL (has scheme and host), use as-is
	if u.IsAbs() {
		return c.addQueryParams(u, queryParams), nil
	}

	// For relative paths, ensure we have a base URL
	if c.baseURL == nil {
		return nil, errors.New("cannot resolve relative path without a base URL")
	}

	// Resolve against base URL
	resolved := c.baseURL.ResolveReference(u)

	return c.addQueryParams(resolved, queryParams), nil
}

// Helper function to add query parameters
func (c *Client) addQueryParams(u *url.URL, params url.Values) *url.URL {
	if params == nil {
		return u
	}

	q := u.Query()
	for k, values := range params {
		for _, v := range values {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	return u
}
