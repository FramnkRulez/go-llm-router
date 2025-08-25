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

// ToolCall represents a tool call request from the LLM
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function details in a tool call
type ToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResult represents the result of executing a tool call
type ToolCallResult struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// Tool represents a tool definition that can be called by the LLM
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a function definition for a tool
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// QueryOptions holds options for LLM queries including tool calls
type QueryOptions struct {
	Temperature float64 `json:"temperature"`
	ForceModel  string  `json:"force_model,omitempty"`
	Tools       []Tool  `json:"tools,omitempty"`
	ToolChoice  string  `json:"tool_choice,omitempty"` // "auto", "none", or specific tool name
}

// QueryResult represents the result of an LLM query
type QueryResult struct {
	Content      string     `json:"content"`
	Model        string     `json:"model"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"`
}

// Provider interface for LLM providers
type Provider interface {
	// Legacy Query method for backward compatibility
	Query(ctx context.Context, messages []Message, temperature float64, forceModel string) (string, string, error)

	// New QueryWithOptions method that supports tool calls
	QueryWithOptions(ctx context.Context, messages []Message, options QueryOptions) (*QueryResult, error)

	HasRemainingRequests(ctx context.Context) bool
	HasRemainingRequestsPerMinute(ctx context.Context) bool
	HasRemainingTokensPerMinute(ctx context.Context, estimatedTokens int) bool
	GetRank() int
	Close()

	// Name returns the name of the provider for error reporting
	Name() string
}

// Config holds common configuration for providers
type Config struct {
	APIKey               string
	Models               []string
	MaxDailyRequests     int
	MaxRequestsPerMinute int
	MaxTokensPerMinute   int
	Rank                 int // Higher rank = higher priority (0 is lowest)
	Timeout              time.Duration
	HTTPClient           httpclient.Client
}
