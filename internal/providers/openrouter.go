package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/FramnkRulez/go-llm-router/internal/httpclient"
	"github.com/FramnkRulez/go-llm-router/provider"
)

// OpenRouterProvider implements the Provider interface for OpenRouter API
type OpenRouterProvider struct {
	apiKey               string
	url                  string
	timeout              time.Duration
	client               httpclient.Client
	models               []string
	referer              string
	xTitle               string
	maxDailyRequests     int
	maxRequestsPerMinute int
	maxTokensPerMinute   int
	rank                 int
	requestsToday        int
	requestsThisMinute   int
	tokensThisMinute     int
	lastReset            time.Time
	lastMinuteReset      time.Time
}

var _ provider.Provider = (*OpenRouterProvider)(nil)

// newOpenRouterProvider creates a new OpenRouter provider
func newOpenRouterProvider(apiKey string, url string, timeout time.Duration, models []string, referer string, xTitle string, httpClient httpclient.Client, maxDailyRequests, maxRequestsPerMinute, maxTokensPerMinute, rank int) (provider.Provider, error) {
	now := time.Now()
	return &OpenRouterProvider{
		url:                  url,
		apiKey:               apiKey,
		timeout:              timeout,
		models:               models,
		client:               httpClient,
		referer:              referer,
		xTitle:               xTitle,
		maxDailyRequests:     maxDailyRequests,
		maxRequestsPerMinute: maxRequestsPerMinute,
		maxTokensPerMinute:   maxTokensPerMinute,
		rank:                 rank,
		requestsToday:        0,
		requestsThisMinute:   0,
		tokensThisMinute:     0,
		lastReset:            now.Truncate(24 * time.Hour),
		lastMinuteReset:      now.Truncate(time.Minute),
	}, nil
}

// Query sends a prompt to OpenRouter and returns the response (legacy method)
func (o *OpenRouterProvider) Query(ctx context.Context, messages []provider.Message, temperature float64, forceModel string) (string, string, error) {
	options := provider.QueryOptions{
		Temperature: temperature,
		ForceModel:  forceModel,
	}

	result, err := o.QueryWithOptions(ctx, messages, options)
	if err != nil {
		return "", "", err
	}

	return result.Content, result.Model, nil
}

// QueryWithOptions sends a prompt to OpenRouter with advanced options including tool calls
func (o *OpenRouterProvider) QueryWithOptions(ctx context.Context, messages []provider.Message, options provider.QueryOptions) (*provider.QueryResult, error) {
	var outerErr error

	if time.Since(o.lastReset) > 24*time.Hour {
		o.requestsToday = 0
		o.lastReset = time.Now().Truncate(24 * time.Hour)
	}

	modelsToUse := o.models
	if options.ForceModel != "" {
		modelsToUse = []string{options.ForceModel}
	}

	for _, model := range modelsToUse {
		// Convert messages to OpenRouter format with file support
		openRouterMessages := make([]map[string]interface{}, 0, len(messages))
		for _, message := range messages {
			msg := map[string]interface{}{
				"role": message.Role,
			}

			// Handle content and files
			if len(message.Files) > 0 {
				// If we have files, we need to use the content array format
				content := make([]map[string]interface{}, 0)

				// Add text content if present
				if message.Content != "" {
					content = append(content, map[string]interface{}{
						"type": "text",
						"text": message.Content,
					})
				}

				// Add file attachments
				for _, file := range message.Files {
					fileContent := map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url": fmt.Sprintf("data:%s;base64,%s", file.MimeType, base64.StdEncoding.EncodeToString(file.Data)),
						},
					}
					content = append(content, fileContent)
				}

				msg["content"] = content
			} else {
				// Simple text message
				msg["content"] = message.Content
			}

			openRouterMessages = append(openRouterMessages, msg)
		}

		requestBody := map[string]interface{}{
			"model":       model,
			"messages":    openRouterMessages,
			"temperature": options.Temperature,
		}

		// Add tools if provided
		if len(options.Tools) > 0 {
			requestBody["tools"] = options.Tools
		}

		// Add tool_choice if provided
		if options.ToolChoice != "" {
			requestBody["tool_choice"] = options.ToolChoice
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

		// Update rate limiting counters
		o.requestsToday++
		o.requestsThisMinute++

		// Estimate tokens for this request (rough approximation)
		estimatedTokens := estimateTokensForMessages(messages)
		o.tokensThisMinute += estimatedTokens

		var result struct {
			Choices []struct {
				Message struct {
					Content   string              `json:"content"`
					ToolCalls []provider.ToolCall `json:"tool_calls,omitempty"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
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

		choice := result.Choices[0]
		queryResult := &provider.QueryResult{
			Content:      choice.Message.Content,
			Model:        model,
			ToolCalls:    choice.Message.ToolCalls,
			FinishReason: choice.FinishReason,
		}

		return queryResult, nil
	}

	return nil, outerErr
}

// Close closes the OpenRouter provider
func (o *OpenRouterProvider) Close() {
	// No cleanup needed for HTTP client
}

// HasRemainingRequests checks if the provider has remaining requests
func (o *OpenRouterProvider) HasRemainingRequests(ctx context.Context) bool {
	// Reset daily counter if needed
	if time.Since(o.lastReset) > 24*time.Hour {
		o.requestsToday = 0
		o.lastReset = time.Now().Truncate(24 * time.Hour)
	}
	return o.maxDailyRequests == 0 || o.requestsToday < o.maxDailyRequests
}

// HasRemainingRequestsPerMinute checks if the provider has remaining requests per minute
func (o *OpenRouterProvider) HasRemainingRequestsPerMinute(ctx context.Context) bool {
	// Reset minute counter if needed
	if time.Since(o.lastMinuteReset) > time.Minute {
		o.requestsThisMinute = 0
		o.tokensThisMinute = 0
		o.lastMinuteReset = time.Now().Truncate(time.Minute)
	}
	return o.maxRequestsPerMinute == 0 || o.requestsThisMinute < o.maxRequestsPerMinute
}

// HasRemainingTokensPerMinute checks if the provider has remaining tokens per minute
func (o *OpenRouterProvider) HasRemainingTokensPerMinute(ctx context.Context, estimatedTokens int) bool {
	// Reset minute counter if needed
	if time.Since(o.lastMinuteReset) > time.Minute {
		o.requestsThisMinute = 0
		o.tokensThisMinute = 0
		o.lastMinuteReset = time.Now().Truncate(time.Minute)
	}
	return o.maxTokensPerMinute == 0 || (o.tokensThisMinute+estimatedTokens) <= o.maxTokensPerMinute
}

// GetRank returns the provider's rank
func (o *OpenRouterProvider) GetRank() int {
	return o.rank
}

// Name returns the name of the provider
func (o *OpenRouterProvider) Name() string {
	return "OpenRouter"
}
