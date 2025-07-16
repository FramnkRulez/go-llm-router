package providers

import (
	"context"
	"time"

	"github.com/FramnkRulez/go-llm-router/internal/httpclient"
)

// Message represents a chat message with role and content
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Provider interface for LLM providers
type Provider interface {
	Query(ctx context.Context, messages []Message, temperature float64, forceModel string) (string, string, error)
	HasRemainingRequests(ctx context.Context) bool
	Close()
}

// Config holds common configuration for providers
type Config struct {
	APIKey           string
	Models           []string
	MaxDailyRequests int
	Timeout          time.Duration
	HTTPClient       httpclient.Client
}
