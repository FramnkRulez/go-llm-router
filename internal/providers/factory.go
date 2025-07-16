package providers

import (
	"time"

	"github.com/FramnkRulez/go-llm-router/internal/httpclient"
)

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string, models []string, maxDailyRequests int) (Provider, error) {
	return newGeminiProvider(apiKey, models, maxDailyRequests)
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(apiKey string, url string, timeout time.Duration, models []string, referer string, xTitle string, httpClient httpclient.Client, maxDailyRequests int) (Provider, error) {
	return newOpenRouterProvider(apiKey, url, timeout, models, referer, xTitle, httpClient, maxDailyRequests)
}
