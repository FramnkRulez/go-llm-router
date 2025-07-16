package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/FramnkRulez/go-llm-router/provider"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

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

// newGeminiProvider creates a new Gemini provider
func newGeminiProvider(apiKey string, models []string, maxDailyRequests int) (provider.Provider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
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

// Query sends a prompt to Gemini and returns the response
func (g *GeminiProvider) Query(ctx context.Context, messages []provider.Message, temperature float64, forceModel string) (string, string, error) {
	if time.Since(g.lastReset) > 24*time.Hour {
		g.requestsToday = 0
		g.lastReset = time.Now().Truncate(24 * time.Hour)
	}

	modelsToUse := g.models
	if forceModel != "" {
		modelsToUse = []string{forceModel}
	}

	var err error
	for _, model := range modelsToUse {
		geminiModel := g.client.GenerativeModel(model)

		textPrompts := []genai.Part{}
		for _, message := range messages {
			textPrompts = append(textPrompts, genai.Text(message.Content))
		}

		var resp *genai.GenerateContentResponse
		resp, err = geminiModel.GenerateContent(ctx, textPrompts...)
		if err != nil {
			continue
		}

		g.requestsToday++

		content := ""
		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				switch p := part.(type) {
				case genai.Text:
					content += string(p)
				}
			}
		}

		return content, model, nil
	}

	return "", "", fmt.Errorf("failed to generate content: %w", err)
}

// Close closes the Gemini client
func (g *GeminiProvider) Close() {
	g.client.Close()
}

// HasRemainingRequests checks if the provider has remaining requests
func (g *GeminiProvider) HasRemainingRequests(ctx context.Context) bool {
	return g.requestsToday < g.maxDailyRequests
}
