# go-llm-router
A Go library that routes LLM requests across multiple providers to maximize free tier usage and stay within API quotas.

## What it does
This library helps you manage multiple LLM API providers (like Gemini, OpenRouter, etc.) by automatically routing requests to available providers based on their remaining quota. When one provider hits its daily limit, it will fall back to the next available provider.  It's up to you to provide valid free models to the OpenRouter provider!

## Installation

```
go get github.com/FramnkRulez/go-llm-router
```

## Usage Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	gollmrouter "github.com/FramnkRulez/go-llm-router"
)

func main() {
	// Create Gemini provider
	geminiConfig := gollmrouter.GeminiConfig{
		APIKey:       "your-gemini-api-key",
		Models:       []string{"gemini-pro", "gemini-pro-vision"}, // Multiple models for fallback
		MaxDailyReqs: 100,
	}
	geminiProvider, err := gollmrouter.NewGeminiProvider(geminiConfig)
	if err != nil {
		panic(err)
	}

	// Create OpenRouter provider
	openRouterConfig := gollmrouter.OpenRouterConfig{
		APIKey:       "your-openrouter-api-key",
		URL:          "https://openrouter.ai/api/v1/chat/completions",
		Models:       []string{"openai/gpt-3.5-turbo", "anthropic/claude-3-haiku"}, // Multiple models for fallback
		MaxDailyReqs: 100,
		Referer:      "your-app",
		XTitle:       "your-title",
		Timeout:      30 * time.Second,
	}
	openRouterProvider, err := gollmrouter.NewOpenRouterProvider(openRouterConfig)
	if err != nil {
		panic(err)
	}

	// Create router with providers
	router, err := gollmrouter.NewRouter(geminiProvider, openRouterProvider)
	if err != nil {
		panic(err)
	}
	defer router.Close()

	ctx := context.Background()
	messages := []gollmrouter.Message{
		{Role: "user", Content: "Hello, who are you?"},
	}

	response, model, err := router.Query(ctx, messages, 0.7, "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Model: %s\nResponse: %s\n", model, response)
}
```

## How Fallback Works

The library provides **two levels of fallback** to maximize reliability:

### 1. Provider-Level Fallback (Model Fallback)
Each provider accepts a slice of models and will try them in order until one succeeds:
- If the first model fails (API error, rate limit, etc.), it automatically tries the next model
- This handles temporary issues with specific models or providers
- Example: `[]string{"gemini-pro", "gemini-pro-vision"}` - if `gemini-pro` fails, it tries `gemini-pro-vision`

### 2. Router-Level Fallback (Provider Fallback)
The router tries each provider in order until one succeeds:
- If a provider is out of quota or fails completely, it moves to the next provider
- This handles quota exhaustion and provider outages
- Example: If Gemini is out of requests, it automatically tries OpenRouter

### Fallback Priority
1. **Router tries providers** in the order they were passed to `NewRouter()`
2. **Each provider tries its models** in the order specified in the config
3. **Only moves to next provider** if all models in the current provider fail

## Creating Custom Providers

You can create your own providers by implementing the `providers.Provider` interface:

```go
type MyCustomProvider struct {
    // your implementation
}

func (p *MyCustomProvider) Query(ctx context.Context, messages []providers.Message, temperature float64, forceModel string) (string, string, error) {
    // your implementation
}

func (p *MyCustomProvider) HasRemainingRequests(ctx context.Context) bool {
    // your implementation
}

func (p *MyCustomProvider) Close() {
    // your implementation
}

// Then use it with the router
router, err := gollmrouter.NewRouter(myCustomProvider, geminiProvider)
```

## How it Works
1. **Configuration**: Set up your API keys and daily limits for each provider
2. **Request Routing**: When you make a request, the router checks which providers have remaining quota
3. **Automatic Fallback**: If the first provider fails or is out of quota, it automatically tries the next available provider
4. **Quota Tracking**: Each provider tracks its daily usage and resets at midnight UTC
5. **Seamless Operation**: Your application continues working even as providers hit their limits

## Supported Providers
- **Google Gemini**: Direct API integration with quota management
- **OpenRouter**: OpenAI-compatible API gateway with access to multiple models
- **Extensible**: Easy to add new providers by implementing the Provider interface
