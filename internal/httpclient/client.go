package httpclient

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Client interface for making HTTP requests
type Client interface {
	Do(ctx context.Context, url string, method string, headers map[string]string, body io.Reader, timeout time.Duration) (*http.Response, string, error)
}

// clientImpl implements the Client interface
type clientImpl struct {
	userAgent string
}

// New creates a new HTTP client
func New(userAgent string) Client {
	return &clientImpl{
		userAgent: userAgent,
	}
}

// Do performs an HTTP request and returns the response
func (c *clientImpl) Do(ctx context.Context, url string, method string, headers map[string]string, body io.Reader, timeout time.Duration) (*http.Response, string, error) {
	finalURL := url

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			finalURL = req.URL.String()
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, finalURL, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, finalURL, err
	}
	return resp, finalURL, nil
}
