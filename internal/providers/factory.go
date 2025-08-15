package providers

import (
	"time"

	"github.com/FramnkRulez/go-llm-router/internal/httpclient"
	"github.com/FramnkRulez/go-llm-router/provider"
)

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string, models []string, maxDailyRequests int) (provider.Provider, error) {
	return newGeminiProvider(apiKey, models, maxDailyRequests)
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(apiKey string, url string, timeout time.Duration, models []string, referer string, xTitle string, httpClient httpclient.Client, maxDailyRequests int) (provider.Provider, error) {
	return newOpenRouterProvider(apiKey, url, timeout, models, referer, xTitle, httpClient, maxDailyRequests)
}

// NewFunctionCallingProvider creates a new function calling provider for LLM APIs that support function calling
func NewFunctionCallingProvider(apiKey string, url string, timeout time.Duration, models []string, httpClient httpclient.Client, maxDailyRequests int, toolExecutor ToolExecutor) (provider.Provider, error) {
	return newFunctionCallingProvider(apiKey, url, timeout, models, httpClient, maxDailyRequests, toolExecutor)
}
