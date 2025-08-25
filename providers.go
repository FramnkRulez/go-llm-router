package gollmrouter

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FramnkRulez/go-llm-router/internal/httpclient"
	"github.com/FramnkRulez/go-llm-router/internal/providers"
	"github.com/FramnkRulez/go-llm-router/provider"
)

// Common API endpoints
const (
	OpenRouterAPIEndpoint = "https://openrouter.ai/api/v1/chat/completions"
)

// FileAttachment represents a file attachment with metadata and data
type FileAttachment = provider.File

// Message represents a chat message with role, content, and optional file attachments
type Message = provider.Message

// ToolCall represents a tool call request from the LLM
type ToolCall = provider.ToolCall

// ToolCallFunction represents the function details in a tool call
type ToolCallFunction = provider.ToolCallFunction

// ToolCallResult represents the result of executing a tool call
type ToolCallResult = provider.ToolCallResult

// Tool represents a tool definition that can be called by the LLM
type Tool = provider.Tool

// ToolFunction represents a function definition for a tool
type ToolFunction = provider.ToolFunction

// QueryOptions holds options for LLM queries including tool calls
type QueryOptions = provider.QueryOptions

// QueryResult represents the result of an LLM query
type QueryResult = provider.QueryResult

// ToolExecutor interface for executing tool calls
type ToolExecutor = providers.ToolExecutor

// GeminiConfig holds configuration for creating a Gemini provider
type GeminiConfig struct {
	APIKey               string
	Models               []string
	MaxDailyReqs         int
	MaxRequestsPerMinute int
	MaxTokensPerMinute   int
	Rank                 int
}

// OpenRouterConfig holds configuration for creating an OpenRouter provider
type OpenRouterConfig struct {
	APIKey               string
	Models               []string
	MaxDailyReqs         int
	MaxRequestsPerMinute int
	MaxTokensPerMinute   int
	Rank                 int
	Referer              string
	XTitle               string
	Timeout              time.Duration
}

// FunctionCallingConfig holds configuration for creating a function calling provider
type FunctionCallingConfig struct {
	APIKey               string
	URL                  string
	Models               []string
	MaxDailyReqs         int
	MaxRequestsPerMinute int
	MaxTokensPerMinute   int
	Rank                 int
	Timeout              time.Duration
	ToolExecutor         ToolExecutor
}

// NewGeminiProvider creates a new Gemini provider with the given configuration
func NewGeminiProvider(config GeminiConfig) (provider.Provider, error) {
	return providers.NewGeminiProvider(
		config.APIKey,
		config.Models,
		config.MaxDailyReqs,
		config.MaxRequestsPerMinute,
		config.MaxTokensPerMinute,
		config.Rank,
	)
}

// NewOpenRouterProvider creates a new OpenRouter provider with the given configuration
func NewOpenRouterProvider(config OpenRouterConfig) (provider.Provider, error) {
	httpClient := httpclient.New("go-llm-router/1.0")

	return providers.NewOpenRouterProvider(
		config.APIKey,
		OpenRouterAPIEndpoint,
		config.Timeout,
		config.Models,
		config.Referer,
		config.XTitle,
		httpClient,
		config.MaxDailyReqs,
		config.MaxRequestsPerMinute,
		config.MaxTokensPerMinute,
		config.Rank,
	)
}

// NewFunctionCallingProvider creates a new function calling provider with the given configuration
func NewFunctionCallingProvider(config FunctionCallingConfig) (provider.Provider, error) {
	httpClient := httpclient.New("go-llm-router/1.0")

	return providers.NewFunctionCallingProvider(
		config.APIKey,
		config.URL,
		config.Timeout,
		config.Models,
		httpClient,
		config.MaxDailyReqs,
		config.MaxRequestsPerMinute,
		config.MaxTokensPerMinute,
		config.Rank,
		config.ToolExecutor,
	)
}

// NewTool creates a new tool definition
func NewTool(name, description string, parameters map[string]interface{}) Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
}

// NewToolCall creates a new tool call
func NewToolCall(id, functionName string, arguments map[string]interface{}) ToolCall {
	return ToolCall{
		ID:   id,
		Type: "function",
		Function: ToolCallFunction{
			Name:      functionName,
			Arguments: arguments,
		},
	}
}

// NewToolCallResult creates a new tool call result
func NewToolCallResult(id string, content interface{}) *ToolCallResult {
	return &ToolCallResult{
		ID:      id,
		Type:    "function",
		Content: content,
	}
}

// NewFileAttachment creates a new file attachment from file data
func NewFileAttachment(fileType, mimeType, name string, data []byte) FileAttachment {
	return FileAttachment{
		Type:     fileType,
		Data:     data,
		MimeType: mimeType,
		Name:     name,
	}
}

// NewFileAttachmentFromPath creates a file attachment from a file path
func NewFileAttachmentFromPath(filePath string) (FileAttachment, error) {
	// Read file data
	file, err := os.Open(filePath)
	if err != nil {
		return FileAttachment{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return FileAttachment{}, err
	}

	// Determine file type and MIME type
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Determine file type based on MIME type
	fileType := "document"
	if strings.HasPrefix(mimeType, "image/") {
		fileType = "image"
	}

	return FileAttachment{
		Type:     fileType,
		Data:     data,
		MimeType: mimeType,
		Name:     filepath.Base(filePath),
	}, nil
}

// NewMessage creates a new message with optional file attachments
func NewMessage(role, content string, files ...FileAttachment) Message {
	return Message{
		Role:    role,
		Content: content,
		Files:   files,
	}
}

// NewMessageWithImage creates a message with an image attachment
func NewMessageWithImage(role, content, imagePath string) (Message, error) {
	imageFile, err := NewFileAttachmentFromPath(imagePath)
	if err != nil {
		return Message{}, err
	}
	return NewMessage(role, content, imageFile), nil
}

// Gemini-specific message creation helpers

// NewGeminiUserMessage creates a user message for Gemini (default for all questions)
func NewGeminiUserMessage(content string, files ...FileAttachment) Message {
	return Message{
		Role:    "user",
		Content: content,
		Files:   files,
	}
}

// NewGeminiModelMessage creates a model/assistant message for Gemini
func NewGeminiModelMessage(content string) Message {
	return Message{
		Role:    "assistant",
		Content: content,
		Files:   []FileAttachment{},
	}
}

// NewGeminiUserMessageWithImage creates a Gemini user message with an image attachment
func NewGeminiUserMessageWithImage(content, imagePath string) (Message, error) {
	imageFile, err := NewFileAttachmentFromPath(imagePath)
	if err != nil {
		return Message{}, err
	}
	return NewGeminiUserMessage(content, imageFile), nil
}

// ValidateGeminiMessage validates that a message has a valid role for Gemini
func ValidateGeminiMessage(message Message) error {
	validRoles := map[string]bool{
		"user":      true,
		"assistant": true,
		"system":    true, // Will be converted to "user" by Gemini
	}

	if !validRoles[message.Role] {
		return fmt.Errorf("invalid role for Gemini: %s (only 'user', 'assistant', and 'system' are supported)", message.Role)
	}

	return nil
}

// ValidateGeminiMessages validates a slice of messages for Gemini compatibility
func ValidateGeminiMessages(messages []Message) error {
	for i, message := range messages {
		if err := ValidateGeminiMessage(message); err != nil {
			return fmt.Errorf("message %d: %w", i, err)
		}
	}
	return nil
}
