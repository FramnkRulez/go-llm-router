package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/FramnkRulez/go-llm-router/provider"
	"google.golang.org/genai"
)

// GeminiRole represents the allowed roles for Gemini API
type GeminiRole string

const (
	// GeminiRoleUser represents a user message (default for all questions)
	GeminiRoleUser GeminiRole = "user"
	// GeminiRoleModel represents a model/assistant response
	GeminiRoleModel GeminiRole = "model"
)

// String returns the string representation of the role
func (r GeminiRole) String() string {
	return string(r)
}

// IsValid checks if the role is valid for Gemini
func (r GeminiRole) IsValid() bool {
	return r == GeminiRoleUser || r == GeminiRoleModel
}

// GeminiProvider implements the Provider interface for Google's Gemini API
type GeminiProvider struct {
	apiKey           string
	client           *genai.Client
	models           []string
	maxDailyRequests int
	requestsToday    int
	lastReset        time.Time
}

var _ provider.Provider = (*GeminiProvider)(nil)

// convertRoleToGemini converts standard chat roles to Gemini-compatible roles
// Returns a strongly typed GeminiRole
func convertRoleToGemini(role string) GeminiRole {
	switch role {
	case "system":
		// Gemini doesn't support "system" role, convert to "user"
		return GeminiRoleUser
	case "user":
		return GeminiRoleUser
	case "assistant":
		return GeminiRoleModel
	default:
		// Default to user for unknown roles
		return GeminiRoleUser
	}
}

// validateGeminiRole validates that a role is acceptable for Gemini
func validateGeminiRole(role string) error {
	geminiRole := convertRoleToGemini(role)
	if !geminiRole.IsValid() {
		return fmt.Errorf("invalid role for Gemini: %s (only 'user' and 'assistant' are supported)", role)
	}
	return nil
}

// newGeminiProvider creates a new Gemini provider
func newGeminiProvider(apiKey string, models []string, maxDailyRequests int) (provider.Provider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		apiKey:           apiKey,
		client:           client,
		models:           models,
		maxDailyRequests: maxDailyRequests,
		requestsToday:    0,
		lastReset:        time.Now().Truncate(24 * time.Hour),
	}, nil
}

// Query sends a prompt to Gemini and returns the response (legacy method)
func (g *GeminiProvider) Query(ctx context.Context, messages []provider.Message, temperature float64, forceModel string) (string, string, error) {
	options := provider.QueryOptions{
		Temperature: temperature,
		ForceModel:  forceModel,
	}

	result, err := g.QueryWithOptions(ctx, messages, options)
	if err != nil {
		return "", "", err
	}

	return result.Content, result.Model, nil
}

// QueryWithOptions sends a prompt to Gemini with advanced options including function calling
func (g *GeminiProvider) QueryWithOptions(ctx context.Context, messages []provider.Message, options provider.QueryOptions) (*provider.QueryResult, error) {
	if time.Since(g.lastReset) > 24*time.Hour {
		g.requestsToday = 0
		g.lastReset = time.Now().Truncate(24 * time.Hour)
	}

	modelsToUse := g.models
	if options.ForceModel != "" {
		modelsToUse = []string{options.ForceModel}
	}

	var err error
	for _, model := range modelsToUse {
		// Convert messages to Gemini format with support for files
		genaiMessages := make([]*genai.Content, 0, len(messages))
		for _, message := range messages {
			// Validate role for Gemini
			if err := validateGeminiRole(message.Role); err != nil {
				return nil, fmt.Errorf("message validation failed: %w", err)
			}
			parts := make([]*genai.Part, 0)

			// Add text content if present
			if message.Content != "" {
				parts = append(parts, &genai.Part{Text: message.Content})
			}

			// Add file attachments if present
			for _, file := range message.Files {
				switch file.Type {
				case "image":
					// Handle image files
					inlineData := &genai.Blob{
						Data:     file.Data,
						MIMEType: file.MimeType,
					}
					parts = append(parts, &genai.Part{InlineData: inlineData})
				case "document":
					// Handle document files (PDF, etc.)
					// Note: Gemini has limited document support, mainly for images
					// For now, we'll skip document files that aren't images
					if file.MimeType == "application/pdf" {
						// PDFs need special handling - skip for now
						continue
					}
				default:
					// Skip unsupported file types
					continue
				}
			}

			// Convert role to Gemini format
			geminiRole := convertRoleToGemini(message.Role)

			genaiMessages = append(genaiMessages, &genai.Content{
				Parts: parts,
				Role:  geminiRole.String(),
			})
		}

		// Create generation config
		config := &genai.GenerateContentConfig{}
		if options.Temperature > 0 {
			temp := float32(options.Temperature)
			config.Temperature = &temp
		}

		// Create tools if provided
		if len(options.Tools) > 0 {
			tools := make([]*genai.Tool, 0, len(options.Tools))
			for _, tool := range options.Tools {
				// Convert parameters to Schema if needed
				// For now, we'll skip complex parameter conversion
				genaiTool := &genai.Tool{
					FunctionDeclarations: []*genai.FunctionDeclaration{
						{
							Name:        tool.Function.Name,
							Description: tool.Function.Description,
							// Parameters field is not directly available in the new API
							// We'll need to handle this differently
						},
					},
				}
				tools = append(tools, genaiTool)
			}
			config.Tools = tools
		}

		// Make the request
		resp, err := g.client.Models.GenerateContent(ctx, model, genaiMessages, config)
		if err != nil {
			continue
		}

		g.requestsToday++

		content := ""
		finishReason := "stop"
		var toolCalls []provider.ToolCall

		for _, candidate := range resp.Candidates {
			if candidate.FinishReason != "" {
				finishReason = string(candidate.FinishReason)
			}

			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					content += part.Text
				}

				// Handle function calls if present
				if part.FunctionCall != nil {
					toolCall := provider.ToolCall{
						ID:   part.FunctionCall.Name, // Use function name as ID
						Type: "function",
						Function: provider.ToolCallFunction{
							Name:      part.FunctionCall.Name,
							Arguments: part.FunctionCall.Args,
						},
					}
					toolCalls = append(toolCalls, toolCall)
				}
			}
		}

		// Note: Gemini now supports function calling with the new SDK
		result := &provider.QueryResult{
			Content:      content,
			Model:        model,
			ToolCalls:    toolCalls,
			FinishReason: finishReason,
		}

		return result, nil
	}

	return nil, fmt.Errorf("failed to generate content: %w", err)
}

// Close closes the Gemini client
func (g *GeminiProvider) Close() {
	// The new genai client doesn't have a Close method
	// Resources are managed automatically
}

// HasRemainingRequests checks if the provider has remaining requests
func (g *GeminiProvider) HasRemainingRequests(ctx context.Context) bool {
	return g.requestsToday < g.maxDailyRequests
}

// Name returns the name of the provider
func (g *GeminiProvider) Name() string {
	return "Gemini"
}
