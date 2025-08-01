// Package gollmrouter provides a simple router for calling different LLM APIs
// to stay within quota limits by rotating through available providers.
//
// The router supports automatic fallback between providers and models,
// making it easy to maximize usage of free tiers across multiple LLM services.
package gollmrouter

import (
	"context"
	"fmt"

	"github.com/FramnkRulez/go-llm-router/provider"
)

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

	for _, provider := range r.providers {
		if !provider.HasRemainingRequests(ctx) {
			continue
		}

		result, model, err := provider.Query(ctx, providerMessages, temperature, forceModel)
		if err == nil {
			return result, model, nil
		}
	}
	return "", "", fmt.Errorf("no provider available or all providers failed")
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
