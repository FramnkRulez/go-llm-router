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
	"github.com/FramnkRulez/go-llm-router/provider"
)

// FunctionCallingProvider implements the Provider interface for LLM APIs that support function calling
// This provider can work with OpenAI, Anthropic, or any other LLM API that supports the OpenAI function calling format
type FunctionCallingProvider struct {
	apiKey               string
	url                  string
	timeout              time.Duration
	client               httpclient.Client
	models               []string
	maxDailyRequests     int
	maxRequestsPerMinute int
	maxTokensPerMinute   int
	rank                 int
	requestsToday        int
	requestsThisMinute   int
	tokensThisMinute     int
	lastReset            time.Time
	lastMinuteReset      time.Time
	toolExecutor         ToolExecutor
}

// ToolExecutor interface for executing tool calls
type ToolExecutor interface {
	ExecuteTool(ctx context.Context, toolCall provider.ToolCall) (*provider.ToolCallResult, error)
	GetAvailableTools() []provider.Tool
}

var _ provider.Provider = (*FunctionCallingProvider)(nil)

// newFunctionCallingProvider creates a new function calling provider
func newFunctionCallingProvider(apiKey string, url string, timeout time.Duration, models []string, httpClient httpclient.Client, maxDailyRequests, maxRequestsPerMinute, maxTokensPerMinute, rank int, toolExecutor ToolExecutor) (provider.Provider, error) {
	now := time.Now()
	return &FunctionCallingProvider{
		url:                  url,
		apiKey:               apiKey,
		timeout:              timeout,
		models:               models,
		client:               httpClient,
		maxDailyRequests:     maxDailyRequests,
		maxRequestsPerMinute: maxRequestsPerMinute,
		maxTokensPerMinute:   maxTokensPerMinute,
		rank:                 rank,
		requestsToday:        0,
		requestsThisMinute:   0,
		tokensThisMinute:     0,
		lastReset:            now.Truncate(24 * time.Hour),
		lastMinuteReset:      now.Truncate(time.Minute),
		toolExecutor:         toolExecutor,
	}, nil
}

// Query sends a prompt to the LLM API and returns the response (legacy method)
func (f *FunctionCallingProvider) Query(ctx context.Context, messages []provider.Message, temperature float64, forceModel string) (string, string, error) {
	options := provider.QueryOptions{
		Temperature: temperature,
		ForceModel:  forceModel,
	}

	result, err := f.QueryWithOptions(ctx, messages, options)
	if err != nil {
		return "", "", err
	}

	return result.Content, result.Model, nil
}

// QueryWithOptions sends a prompt to the LLM API with advanced options including function calling
func (f *FunctionCallingProvider) QueryWithOptions(ctx context.Context, messages []provider.Message, options provider.QueryOptions) (*provider.QueryResult, error) {
	if time.Since(f.lastReset) > 24*time.Hour {
		f.requestsToday = 0
		f.lastReset = time.Now().Truncate(24 * time.Hour)
	}

	modelsToUse := f.models
	if options.ForceModel != "" {
		modelsToUse = []string{options.ForceModel}
	}

	// If no tools are provided but we have a tool executor, get available tools
	if len(options.Tools) == 0 && f.toolExecutor != nil {
		options.Tools = f.toolExecutor.GetAvailableTools()
	}

	var outerErr error
	for _, model := range modelsToUse {
		// Convert messages to API format
		apiMessages := make([]map[string]interface{}, 0, len(messages))
		for _, message := range messages {
			msg := map[string]interface{}{
				"role":    message.Role,
				"content": message.Content,
			}

			// Add file attachments if present
			if len(message.Files) > 0 {
				files := make([]map[string]interface{}, 0, len(message.Files))
				for _, file := range message.Files {
					fileData := map[string]interface{}{
						"type":      file.Type,
						"mime_type": file.MimeType,
						"name":      file.Name,
						"data":      file.Data,
					}
					files = append(files, fileData)
				}
				msg["files"] = files
			}

			apiMessages = append(apiMessages, msg)
		}

		requestBody := map[string]interface{}{
			"model":       model,
			"messages":    apiMessages,
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

		// Make the initial request
		result, err := f.makeRequest(ctx, requestBody)
		if err != nil {
			outerErr = err
			continue
		}

		// Handle tool calls if present
		if len(result.ToolCalls) > 0 && f.toolExecutor != nil {
			// Execute tool calls
			toolResults := make([]provider.ToolCallResult, 0, len(result.ToolCalls))
			for _, toolCall := range result.ToolCalls {
				toolResult, err := f.toolExecutor.ExecuteTool(ctx, toolCall)
				if err != nil {
					// Log error but continue with other tool calls
					fmt.Printf("Tool execution failed for %s: %v\n", toolCall.Function.Name, err)
					continue
				}
				toolResults = append(toolResults, *toolResult)
			}

			// Add tool results to messages and make another request
			if len(toolResults) > 0 {
				// Create a new message with tool results
				toolMessage := map[string]interface{}{
					"role":         "tool",
					"tool_results": toolResults,
				}

				// Add tool results to the conversation
				updatedMessages := make([]map[string]interface{}, len(apiMessages)+1)
				copy(updatedMessages, apiMessages)
				updatedMessages[len(apiMessages)] = toolMessage

				// Make another request with tool results
				requestBody["messages"] = updatedMessages
				finalResult, err := f.makeRequest(ctx, requestBody)
				if err != nil {
					outerErr = err
					continue
				}

				return finalResult, nil
			}
		}

		return result, nil
	}

	return nil, outerErr
}

// makeRequest makes a single request to the LLM API
func (f *FunctionCallingProvider) makeRequest(ctx context.Context, requestBody map[string]interface{}) (*provider.QueryResult, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, _, err := f.client.Do(ctx, f.url, "POST", map[string]string{
		"Authorization": "Bearer " + f.apiKey,
		"Content-Type":  "application/json",
	}, bytes.NewBuffer(jsonData), f.timeout)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Update rate limiting counters
	f.requestsToday++
	f.requestsThisMinute++

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
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response choices received")
	}

	choice := result.Choices[0]
	queryResult := &provider.QueryResult{
		Content:      choice.Message.Content,
		Model:        requestBody["model"].(string),
		ToolCalls:    choice.Message.ToolCalls,
		FinishReason: choice.FinishReason,
	}

	return queryResult, nil
}

// Close closes the function calling provider
func (f *FunctionCallingProvider) Close() {
	// No cleanup needed for HTTP client
}

// HasRemainingRequests checks if the provider has remaining requests
func (f *FunctionCallingProvider) HasRemainingRequests(ctx context.Context) bool {
	// Reset daily counter if needed
	if time.Since(f.lastReset) > 24*time.Hour {
		f.requestsToday = 0
		f.lastReset = time.Now().Truncate(24 * time.Hour)
	}
	return f.maxDailyRequests == 0 || f.requestsToday < f.maxDailyRequests
}

// HasRemainingRequestsPerMinute checks if the provider has remaining requests per minute
func (f *FunctionCallingProvider) HasRemainingRequestsPerMinute(ctx context.Context) bool {
	// Reset minute counter if needed
	if time.Since(f.lastMinuteReset) > time.Minute {
		f.requestsThisMinute = 0
		f.tokensThisMinute = 0
		f.lastMinuteReset = time.Now().Truncate(time.Minute)
	}
	return f.maxRequestsPerMinute == 0 || f.requestsThisMinute < f.maxRequestsPerMinute
}

// HasRemainingTokensPerMinute checks if the provider has remaining tokens per minute
func (f *FunctionCallingProvider) HasRemainingTokensPerMinute(ctx context.Context, estimatedTokens int) bool {
	// Reset minute counter if needed
	if time.Since(f.lastMinuteReset) > time.Minute {
		f.requestsThisMinute = 0
		f.tokensThisMinute = 0
		f.lastMinuteReset = time.Now().Truncate(time.Minute)
	}
	return f.maxTokensPerMinute == 0 || (f.tokensThisMinute+estimatedTokens) <= f.maxTokensPerMinute
}

// GetRank returns the provider's rank
func (f *FunctionCallingProvider) GetRank() int {
	return f.rank
}

// Name returns the name of the provider
func (f *FunctionCallingProvider) Name() string {
	return "FunctionCalling"
}
