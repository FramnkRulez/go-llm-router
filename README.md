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
	config := gollmrouter.Config{
		GeminiAPIKey:       "your-gemini-api-key",
		GeminiModels:       []string{"gemini-pro"},
		GeminiMaxDailyReqs: 100,

		OpenRouterAPIKey:       "your-openrouter-api-key",
		OpenRouterURL:          "https://openrouter.ai/api/v1/chat/completions",
		OpenRouterModels:       []string{"openai/gpt-3.5-turbo"},
		OpenRouterMaxDailyReqs: 100,
		OpenRouterReferer:      "your-app",
		OpenRouterXTitle:       "your-title",
		OpenRouterTimeout:      30 * time.Second,

		HTTPTimeout: 30 * time.Second,
		UserAgent:   "go-llm-router/1.0",
	}

	router, err := gollmrouter.NewRouter(config)
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
