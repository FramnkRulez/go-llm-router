package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/FramnkRulez/go-llm-router/internal/httpclient"
)

// OpenRouterProvider implements the Provider interface for OpenRouter API
type OpenRouterProvider struct {
	apiKey           string
	url              string
	timeout          time.Duration
	client           httpclient.Client
	models           []string
	referer          string
	xTitle           string
	maxDailyRequests int
	requestsToday    int
	lastReset        time.Time
}

var _ Provider = (*OpenRouterProvider)(nil)

// newOpenRouterProvider creates a new OpenRouter provider
func newOpenRouterProvider(apiKey string, url string, timeout time.Duration, models []string, referer string, xTitle string, httpClient httpclient.Client, maxDailyRequests int) (Provider, error) {
	return &OpenRouterProvider{
		url:              url,
		apiKey:           apiKey,
		timeout:          timeout,
		models:           models,
		client:           httpClient,
		referer:          referer,
		xTitle:           xTitle,
		maxDailyRequests: maxDailyRequests,
		requestsToday:    0,
		lastReset:        time.Now().Truncate(24 * time.Hour),
	}, nil
}

// Query sends a prompt to OpenRouter and returns the response
func (o *OpenRouterProvider) Query(ctx context.Context, messages []Message, temperature float64, forceModel string) (string, string, error) {
	var outerErr error

	if time.Since(o.lastReset) > 24*time.Hour {
		o.requestsToday = 0
		o.lastReset = time.Now().Truncate(24 * time.Hour)
	}

	modelsToUse := o.models
	if forceModel != "" {
		modelsToUse = []string{forceModel}
	}

	for _, model := range modelsToUse {
		requestBody := map[string]interface{}{
			"model":       model,
			"messages":    messages,
			"temperature": temperature,
		}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			outerErr = fmt.Errorf("failed to marshal request: %w", err)
			continue
		}

		resp, _, err := o.client.Do(ctx, o.url, "POST", map[string]string{
			"Authorization": "Bearer " + o.apiKey,
			"Content-Type":  "application/json",
			"HTTP-Referer":  o.referer,
			"X-Title":       o.xTitle,
		}, bytes.NewBuffer(jsonData), o.timeout)

		if err != nil {
			outerErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			outerErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			outerErr = fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
			continue
		}

		o.requestsToday++

		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			outerErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if len(result.Choices) == 0 {
			outerErr = fmt.Errorf("no response choices received")
			continue
		}

		return result.Choices[0].Message.Content, model, nil
	}

	return "", "", outerErr
}

// Close closes the OpenRouter provider
func (o *OpenRouterProvider) Close() {
	// No cleanup needed for HTTP client
}

// HasRemainingRequests checks if the provider has remaining requests
func (o *OpenRouterProvider) HasRemainingRequests(ctx context.Context) bool {
	return o.requestsToday < o.maxDailyRequests
}
