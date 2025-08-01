# go-llm-router

[![Go Reference](https://pkg.go.dev/badge/github.com/FramnkRulez/go-llm-router.svg)](https://pkg.go.dev/github.com/FramnkRulez/go-llm-router)
[![Go Report Card](https://goreportcard.com/badge/github.com/FramnkRulez/go-llm-router)](https://goreportcard.com/report/github.com/FramnkRulez/go-llm-router)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library that routes LLM requests across multiple providers to maximize free tier usage and stay within API quotas. **Now with file attachment support!**

## What it does
This library helps you manage multiple LLM API providers (like Gemini, OpenRouter, etc.) by automatically routing requests to available providers based on their remaining quota. When one provider hits its daily limit, it will fall back to the next available provider. It's up to you to provide valid models to each provider.

**New Feature**: File attachment support for images and documents, allowing you to send files along with your text messages to supported models.

## Installation

```bash
go get github.com/FramnkRulez/go-llm-router
```

## Quick Start

### Basic Text-Only Usage

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
	geminiProvider, _ := gollmrouter.NewGeminiProvider(gollmrouter.GeminiConfig{
		APIKey:       "your-gemini-api-key",
		Models:       []string{"gemini-pro", "gemini-pro-vision"},
		MaxDailyReqs: 100,
	})

	// Create OpenRouter provider
	openRouterProvider, _ := gollmrouter.NewOpenRouterProvider(gollmrouter.OpenRouterConfig{
		APIKey:       "your-openrouter-api-key",
		URL:          "https://openrouter.ai/api/v1/chat/completions",
		Models:       []string{"openai/gpt-3.5-turbo", "anthropic/claude-3-haiku"},
		MaxDailyReqs: 100,
		Referer:      "your-app",
		XTitle:       "your-title",
		Timeout:      30 * time.Second,
	})

	// Create router with providers
	router, _ := gollmrouter.NewRouter(geminiProvider, openRouterProvider)
	defer router.Close()

	// Use the router
	response, model, _ := router.Query(context.Background(), []gollmrouter.Message{
		{Role: "user", Content: "Hello, who are you?"},
	}, 0.7, "")

	fmt.Printf("Model: %s\nResponse: %s\n", model, response)
}
```

### File Attachment Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	gollmrouter "github.com/FramnkRulez/go-llm-router"
)

func main() {
	// Create providers that support file attachments
	geminiProvider, _ := gollmrouter.NewGeminiProvider(gollmrouter.GeminiConfig{
		APIKey:       "your-gemini-api-key",
		Models:       []string{"gemini-pro-vision", "gemini-pro"}, // Vision models support images
		MaxDailyReqs: 100,
	})

	openRouterProvider, _ := gollmrouter.NewOpenRouterProvider(gollmrouter.OpenRouterConfig{
		APIKey:       "your-openrouter-api-key",
		URL:          "https://openrouter.ai/api/v1/chat/completions",
		Models:       []string{"openai/gpt-4-vision-preview", "anthropic/claude-3-haiku"},
		MaxDailyReqs: 100,
		Referer:      "your-app",
		XTitle:       "your-title",
		Timeout:      30 * time.Second,
	})

	router, _ := gollmrouter.NewRouter(geminiProvider, openRouterProvider)
	defer router.Close()

	ctx := context.Background()

	// Example 1: Message with image attachment
	imageMessage, err := gollmrouter.NewMessageWithImage("user", "What do you see in this image?", "path/to/image.jpg")
	if err != nil {
		log.Fatal(err)
	}

	response, model, err := router.Query(ctx, []gollmrouter.Message{imageMessage}, 0.7, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Model: %s\nResponse: %s\n", model, response)

	// Example 2: Message with multiple file attachments
	imageFile1, _ := gollmrouter.NewFileAttachmentFromPath("path/to/image1.jpg")
	imageFile2, _ := gollmrouter.NewFileAttachmentFromPath("path/to/image2.png")
	
	multiFileMessage := gollmrouter.NewMessage("user", "Compare these two images:", imageFile1, imageFile2)
	response, model, err = router.Query(ctx, []gollmrouter.Message{multiFileMessage}, 0.7, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Model: %s\nResponse: %s\n", model, response)
}
```

## File Attachment Features

### Supported File Types

- **Images**: JPEG, PNG, WebP, HEIC (supported by Gemini and vision-capable models via OpenRouter)
- **Documents**: Limited support for PDFs and other documents (varies by provider)

### File Attachment Methods

```go
// Create a file attachment from a file path
file, err := gollmrouter.NewFileAttachmentFromPath("path/to/image.jpg")

// Create a file attachment from bytes (useful for web uploads)
file := gollmrouter.NewFileAttachment("image", "image/jpeg", "image.jpg", imageBytes)

// Create a message with file attachments
message := gollmrouter.NewMessage("user", "Analyze this image", file)

// Create a message with image attachment (convenience method)
message, err := gollmrouter.NewMessageWithImage("user", "What's in this image?", "path/to/image.jpg")
```

### File Attachment Structure

```go
type FileAttachment struct {
    Type     string // "image", "document", etc.
    Data     []byte // file data
    MimeType string // MIME type (e.g., "image/jpeg")
    Name     string // filename
}
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
- If you would prefer to use quota-based fallback instead, simply pass in a copy of the Gemini provider as a new provider to the router with the specified model.

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
6. **File Support**: File attachments are automatically converted to the appropriate format for each provider

## Supported Providers
- **Google Gemini**: Direct API integration with quota management and image support
- **OpenRouter**: OpenAI-compatible API gateway with access to multiple models (check OpenRouter for a list of available free models!)

## Keywords
LLM, AI, Router, Fallback, Quota Management, Gemini, OpenRouter, OpenAI, Claude, API, Go, Golang, Library, File Attachments, Image Analysis, Vision Models

## License
MIT

