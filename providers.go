package gollmrouter

import (
	"time"

	"github.com/FramnkRulez/go-llm-router/internal/httpclient"
	"github.com/FramnkRulez/go-llm-router/internal/providers"
)

// GeminiConfig holds configuration for creating a Gemini provider
type GeminiConfig struct {
	APIKey       string
	Models       []string
	MaxDailyReqs int
}

// OpenRouterConfig holds configuration for creating an OpenRouter provider
type OpenRouterConfig struct {
	APIKey       string
	URL          string
	Models       []string
	MaxDailyReqs int
	Referer      string
	XTitle       string
	Timeout      time.Duration
}

// NewGeminiProvider creates a new Gemini provider with the given configuration
func NewGeminiProvider(config GeminiConfig) (providers.Provider, error) {
	return providers.NewGeminiProvider(
		config.APIKey,
		config.Models,
		config.MaxDailyReqs,
	)
}

// NewOpenRouterProvider creates a new OpenRouter provider with the given configuration
func NewOpenRouterProvider(config OpenRouterConfig) (providers.Provider, error) {
	httpClient := httpclient.New("go-llm-router/1.0")

	return providers.NewOpenRouterProvider(
		config.APIKey,
		config.URL,
		config.Timeout,
		config.Models,
		config.Referer,
		config.XTitle,
		httpClient,
		config.MaxDailyReqs,
	)
}
