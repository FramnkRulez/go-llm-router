// Package go-llm-router provides a simple router for calling different LLM APIs
// to stay within quota limits by rotating through available providers.
package gollmrouter

import (
	"context"
	"fmt"

	"github.com/FramnkRulez/go-llm-router/internal/providers"
)

// Message represents a chat message with role and content
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Router manages multiple LLM providers and routes requests to available ones
type Router struct {
	providers []providers.Provider
}

// NewRouter creates a new router with the specified providers
func NewRouter(providers ...providers.Provider) (*Router, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	return &Router{
		providers: providers,
	}, nil
}

// Query sends a prompt to available LLM providers and returns the first successful response
func (r *Router) Query(ctx context.Context, messages []Message, temperature float64, forceModel string) (string, string, error) {
	// Convert messages to provider format
	providerMessages := make([]providers.Message, len(messages))
	for i, msg := range messages {
		providerMessages[i] = providers.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
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

// HasRemainingRequests checks if any provider has remaining requests
func (r *Router) HasRemainingRequests(ctx context.Context) bool {
	for _, provider := range r.providers {
		if provider.HasRemainingRequests(ctx) {
			return true
		}
	}
	return false
}

// Close closes all providers
func (r *Router) Close() {
	for _, provider := range r.providers {
		provider.Close()
	}
}
