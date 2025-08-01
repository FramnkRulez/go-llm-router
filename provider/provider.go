package provider

import (
	"context"
	"time"

	"github.com/FramnkRulez/go-llm-router/internal/httpclient"
)

// File represents a file attachment with metadata and data
type File struct {
	Type     string `json:"type"`      // "image", "document", etc.
	Data     []byte `json:"data"`      // file data (base64 encoded for JSON)
	MimeType string `json:"mime_type"` // MIME type
	Name     string `json:"name"`      // filename
}

// Message represents a chat message with role, content, and optional file attachments
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Files   []File `json:"files,omitempty"`
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
