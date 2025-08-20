// Package gollmrouter provides a simple router for calling different LLM APIs
// to stay within quota limits by rotating through available providers.
//
// The router supports automatic fallback between providers and models,
// making it easy to maximize usage of free tiers across multiple LLM services.
// It also supports tool calls for MCP (Model Context Protocol) servers.
package gollmrouter

import (
	"context"
	"fmt"
	"strings"

	"github.com/FramnkRulez/go-llm-router/provider"
)

// RouterError represents an error that occurred when all providers failed
type RouterError struct {
	Errors []ProviderError
}

// ProviderError represents an error from a specific provider
type ProviderError struct {
	ProviderName string
	Error        error
}

// Error returns a formatted error message with details from all failed providers
func (r *RouterError) Error() string {
	if len(r.Errors) == 0 {
		return "no provider available or all providers failed"
	}

	var errorMsgs []string
	for _, err := range r.Errors {
		errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %v", err.ProviderName, err.Error))
	}

	return fmt.Sprintf("all providers failed:\n%s", strings.Join(errorMsgs, "\n"))
}

// Unwrap returns the underlying errors
func (r *RouterError) Unwrap() []error {
	var errors []error
	for _, err := range r.Errors {
		errors = append(errors, err.Error)
	}
	return errors
}

// GetErrors returns the list of provider errors
func (r *RouterError) GetErrors() []ProviderError {
	return r.Errors
}

// IsRouterError checks if an error is a RouterError
func IsRouterError(err error) bool {
	_, ok := err.(*RouterError)
	return ok
}

// GetRouterError extracts RouterError from an error if it is one
func GetRouterError(err error) (*RouterError, bool) {
	routerErr, ok := err.(*RouterError)
	return routerErr, ok
}

// Router manages multiple LLM providers and routes requests to available ones.
// It automatically handles fallback between providers based on quota availability
// and request success/failure.
type Router struct {
	providers []provider.Provider
}

// NewRouter creates a new router with the specified providers.
// Providers will be tried in the order they are passed to this function.
// At least one provider must be specified.
func NewRouter(providers ...provider.Provider) (*Router, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	return &Router{
		providers: providers,
	}, nil
}

// Query sends a prompt to available LLM providers and returns the first successful response.
// It automatically handles fallback between providers and models within each provider.
//
// Parameters:
//   - ctx: Context for the request
//   - messages: Array of chat messages to send (can include file attachments)
//   - temperature: Controls randomness (0.0 = deterministic, 1.0 = very random)
//   - forceModel: If specified, forces use of this specific model (optional)
//
// Returns:
//   - response: The generated text response
//   - model: The name of the model that generated the response
//   - error: Any error that occurred (nil if successful)
func (r *Router) Query(ctx context.Context, messages []provider.Message, temperature float64, forceModel string) (string, string, error) {
	// Convert messages to provider format (including file attachments)
	providerMessages := make([]provider.Message, len(messages))
	for i, msg := range messages {
		providerMessages[i] = provider.Message{
			Role:    msg.Role,
			Content: msg.Content,
			Files:   make([]provider.File, len(msg.Files)),
		}
		// Copy file attachments
		copy(providerMessages[i].Files, msg.Files)
	}

	options := provider.QueryOptions{
		Temperature: temperature,
		ForceModel:  forceModel,
	}

	var routerError RouterError

	for i, provider := range r.providers {
		providerName := provider.Name()
		if providerName == "" {
			providerName = fmt.Sprintf("Provider %d", i+1)
		}

		if !provider.HasRemainingRequests(ctx) {
			routerError.Errors = append(routerError.Errors, ProviderError{
				ProviderName: providerName,
				Error:        fmt.Errorf("no remaining requests"),
			})
			continue
		}

		result, err := provider.QueryWithOptions(ctx, providerMessages, options)
		if err == nil {
			return result.Content, result.Model, nil
		}

		// Collect the error
		routerError.Errors = append(routerError.Errors, ProviderError{
			ProviderName: providerName,
			Error:        err,
		})
	}

	if len(routerError.Errors) == 0 {
		return "", "", fmt.Errorf("no providers configured")
	}

	return "", "", &routerError
}

// QueryWithOptions sends a prompt to available LLM providers with advanced options including tool calls.
// It automatically handles fallback between providers and models within each provider.
//
// Parameters:
//   - ctx: Context for the request
//   - messages: Array of chat messages to send (can include file attachments)
//   - options: Query options including temperature, model, tools, and tool choice
//
// Returns:
//   - result: The query result containing content, model, tool calls, and finish reason
//   - error: Any error that occurred (nil if successful)
func (r *Router) QueryWithOptions(ctx context.Context, messages []provider.Message, options provider.QueryOptions) (*provider.QueryResult, error) {
	// Convert messages to provider format (including file attachments)
	providerMessages := make([]provider.Message, len(messages))
	for i, msg := range messages {
		providerMessages[i] = provider.Message{
			Role:    msg.Role,
			Content: msg.Content,
			Files:   make([]provider.File, len(msg.Files)),
		}
		// Copy file attachments
		copy(providerMessages[i].Files, msg.Files)
	}

	var routerError RouterError

	for i, provider := range r.providers {
		providerName := provider.Name()
		if providerName == "" {
			providerName = fmt.Sprintf("Provider %d", i+1)
		}

		if !provider.HasRemainingRequests(ctx) {
			routerError.Errors = append(routerError.Errors, ProviderError{
				ProviderName: providerName,
				Error:        fmt.Errorf("no remaining requests"),
			})
			continue
		}

		result, err := provider.QueryWithOptions(ctx, providerMessages, options)
		if err == nil {
			return result, nil
		}

		// Collect the error
		routerError.Errors = append(routerError.Errors, ProviderError{
			ProviderName: providerName,
			Error:        err,
		})
	}

	if len(routerError.Errors) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	return nil, &routerError
}

// HasRemainingRequests checks if any provider has remaining requests.
// This can be used to check if the router can handle new requests
// before actually making them.
func (r *Router) HasRemainingRequests(ctx context.Context) bool {
	for _, provider := range r.providers {
		if provider.HasRemainingRequests(ctx) {
			return true
		}
	}
	return false
}

// Close closes all providers and releases any resources they hold.
// This should be called when you're done using the router.
func (r *Router) Close() {
	for _, provider := range r.providers {
		provider.Close()
	}
}
