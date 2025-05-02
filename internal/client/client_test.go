package client

import (
	"net/url"
	"strings"
	"testing"
)

func TestClient_ResolveURL(t *testing.T) {
	// Helper to create a base URL for testing
	baseURL, err := url.Parse("https://example.com/api")
	if err != nil {
		t.Fatalf("Failed to parse base URL: %v", err)
	}

	tests := []struct {
		name        string
		client      *Client
		pathOrURL   string
		queryParams url.Values
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "absolute URL",
			client: &Client{
				baseURL: baseURL,
			},
			pathOrURL:   "https://other.com/path",
			queryParams: nil,
			want:        "https://other.com/path",
			wantErr:     false,
		},
		{
			name: "absolute URL with query params",
			client: &Client{
				baseURL: baseURL,
			},
			pathOrURL: "https://other.com/path",
			queryParams: url.Values{
				"key": []string{"value"},
			},
			want:    "https://other.com/path?key=value",
			wantErr: false,
		},
		{
			name: "relative path",
			client: &Client{
				baseURL: baseURL,
			},
			pathOrURL:   "/users",
			queryParams: nil,
			want:        "https://example.com/api/users",
			wantErr:     false,
		},
		{
			name: "relative path with query params",
			client: &Client{
				baseURL: baseURL,
			},
			pathOrURL: "/users",
			queryParams: url.Values{
				"page":  []string{"1"},
				"limit": []string{"10"},
			},
			want:    "https://example.com/api/users?limit=10&page=1",
			wantErr: false,
		},
		{
			name:        "invalid URL",
			client:      &Client{baseURL: baseURL},
			pathOrURL:   ":%invalid",
			queryParams: nil,
			wantErr:     true,
			errContains: "invalid URL or path",
		},
		{
			name:        "relative path without base URL",
			client:      &Client{baseURL: nil},
			pathOrURL:   "/users",
			queryParams: nil,
			wantErr:     true,
			errContains: "cannot resolve relative path without a base URL",
		},
		{
			name: "empty path",
			client: &Client{
				baseURL: baseURL,
			},
			pathOrURL:   "",
			queryParams: nil,
			want:        "https://example.com/api",
			wantErr:     false,
		},
		{
			name: "path starting without slash",
			client: &Client{
				baseURL: baseURL,
			},
			pathOrURL:   "users",
			queryParams: nil,
			want:        "https://example.com/api/users",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.resolveURL(tt.pathOrURL, tt.queryParams)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error message does not contain %q: %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if got.String() != tt.want {
				t.Errorf("Got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestClient_AddQueryParams(t *testing.T) {
	baseURL, err := url.Parse("https://example.com/api")
	if err != nil {
		t.Fatalf("Failed to parse base URL: %v", err)
	}

	client := &Client{}

	tests := []struct {
		name   string
		u      *url.URL
		params url.Values
		want   string
	}{
		{
			name:   "nil params",
			u:      baseURL,
			params: nil,
			want:   "https://example.com/api",
		},
		{
			name:   "empty params",
			u:      baseURL,
			params: url.Values{},
			want:   "https://example.com/api",
		},
		{
			name: "single param",
			u:    baseURL,
			params: url.Values{
				"key": []string{"value"},
			},
			want: "https://example.com/api?key=value",
		},
		{
			name: "multiple params",
			u:    baseURL,
			params: url.Values{
				"page":  []string{"1"},
				"limit": []string{"10"},
			},
			want: "https://example.com/api?limit=10&page=1",
		},
		{
			name: "multiple values for same param",
			u:    baseURL,
			params: url.Values{
				"tag": []string{"golang", "testing", "http"},
			},
			want: "https://example.com/api?tag=golang&tag=testing&tag=http",
		},
		{
			name: "url with existing params",
			u: func() *url.URL {
				u, _ := url.Parse("https://example.com/api?existing=true")
				return u
			}(),
			params: url.Values{
				"new": []string{"param"},
			},
			want: "https://example.com/api?existing=true&new=param",
		},
		{
			name: "params with special characters",
			u:    baseURL,
			params: url.Values{
				"search": []string{"hello world"},
				"filter": []string{"status=active&type=user"},
			},
			want: "https://example.com/api?filter=status%3Dactive%26type%3Duser&search=hello+world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the URL to test that the method doesn't modify unexpected parts
			uCopy := *tt.u

			got := client.addQueryParams(&uCopy, tt.params)

			if got.String() != tt.want {
				t.Errorf("Got %q, want %q", got.String(), tt.want)
			}

			// Verify that it returns the same URL instance that was passed in
			if got != &uCopy {
				t.Errorf("Expected to return the same URL instance")
			}
		})
	}
}
