# go-llm-router

[![Go Reference](https://pkg.go.dev/badge/github.com/FramnkRulez/go-llm-router.svg)](https://pkg.go.dev/github.com/FramnkRulez/go-llm-router)
[![Go Report Card](https://goreportcard.com/badge/github.com/FramnkRulez/go-llm-router)](https://goreportcard.com/report/github.com/FramnkRulez/go-llm-router)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library that routes LLM requests across multiple providers to maximize free tier usage and stay within API quotas. **Now with file attachment support and function calling!**

## What it does
This library helps you manage multiple LLM API providers (like Gemini, OpenRouter, etc.) by automatically routing requests to available providers based on their remaining quota. When one provider hits its daily limit, it will fall back to the next available provider. It's up to you to provide valid models to each provider.

**New Features**:
- **File attachment support** for images and documents, allowing you to send files along with your text messages to supported models
- **Function calling support** for LLM APIs that support function calling (OpenAI, Anthropic, etc.), enabling the LLM to execute custom functions
- **Enhanced Gemini support** using the latest official Google Gen AI Go SDK with full function calling capabilities

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
	// Create Gemini provider (now with function calling support!)
	geminiProvider, _ := gollmrouter.NewGeminiProvider(gollmrouter.GeminiConfig{
		APIKey:       "your-gemini-api-key",
		Models:       []string{"gemini-2.0-flash", "gemini-1.5-flash"}, // Latest models
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
		Models:       []string{"gemini-2.0-flash", "gemini-1.5-flash"}, // Vision models support images
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

	// Create a message with an image attachment
	message, err := gollmrouter.NewMessageWithImage("user", "What's in this image?", "path/to/image.jpg")
	if err != nil {
		log.Fatal(err)
	}

	response, model, _ := router.Query(context.Background(), []gollmrouter.Message{message}, 0.7, "")
	fmt.Printf("Model: %s\nResponse: %s\n", model, response)
}
```

### Function Calling Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	gollmrouter "github.com/FramnkRulez/go-llm-router"
	"github.com/FramnkRulez/go-llm-router/ai"
)

func main() {
	// Create a tool executor with basic tools
	toolExecutor := ai.NewSimpleToolExecutor()

	// Create Gemini provider with function calling support (using the new official SDK!)
	geminiProvider, _ := gollmrouter.NewGeminiProvider(gollmrouter.GeminiConfig{
		APIKey:       "your-gemini-api-key",
		Models:       []string{"gemini-2.0-flash", "gemini-1.5-flash"},
		MaxDailyReqs: 100,
	})

	// Create function calling provider for other LLM APIs
	functionCallingProvider, _ := gollmrouter.NewFunctionCallingProvider(gollmrouter.FunctionCallingConfig{
		APIKey:       "your-openai-api-key",
		URL:          "https://api.openai.com/v1/chat/completions",
		Models:       []string{"gpt-4", "gpt-3.5-turbo"},
		MaxDailyReqs: 100,
		Timeout:      30 * time.Second,
		ToolExecutor: toolExecutor,
	})

	// Create router with all providers
	router, _ := gollmrouter.NewRouter(geminiProvider, functionCallingProvider)
	defer router.Close()

	// Query with function calls
	options := gollmrouter.QueryOptions{
		Temperature: 0.7,
		Tools:       toolExecutor.GetAvailableTools(),
		ToolChoice:  "auto",
	}

	result, err := router.QueryWithOptions(context.Background(), []gollmrouter.Message{
		{Role: "user", Content: "What is the current time and calculate 15 * 23?"},
	}, options)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Model: %s\nContent: %s\n", result.Model, result.Content)
	if len(result.ToolCalls) > 0 {
		fmt.Printf("Function Calls: %+v\n", result.ToolCalls)
	}
}
```

## Advanced Usage

### Creating Custom Tool Executors

You can create custom tool executors to handle specific function calls:

```go
type WeatherToolExecutor struct {
	baseExecutor *ai.SimpleToolExecutor
}

func NewWeatherToolExecutor() *WeatherToolExecutor {
	return &WeatherToolExecutor{
		baseExecutor: ai.NewSimpleToolExecutor(),
	}
}

func (w *WeatherToolExecutor) ExecuteTool(ctx context.Context, toolCall gollmrouter.ToolCall) (*gollmrouter.ToolCallResult, error) {
	// Handle weather tool
	if toolCall.Function.Name == "get_weather" {
		location, ok := toolCall.Function.Arguments["location"].(string)
		if !ok {
			return nil, fmt.Errorf("location argument is required")
		}
		
		// Call weather API and return result
		weatherData := map[string]interface{}{
			"location": location,
			"temperature": "22°C",
			"condition": "Sunny",
		}
		
		return &gollmrouter.ToolCallResult{
			ID:      toolCall.ID,
			Type:    "function",
			Content: weatherData,
		}, nil
	}
	
	// Delegate to base executor for other tools
	return w.baseExecutor.ExecuteTool(ctx, toolCall)
}

func (w *WeatherToolExecutor) GetAvailableTools() []gollmrouter.Tool {
	tools := w.baseExecutor.GetAvailableTools()
	
	// Add weather tool
	weatherTool := gollmrouter.NewTool(
		"get_weather",
		"Get weather information for a location",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"location": map[string]interface{}{
					"type":        "string",
					"description": "City name or location",
				},
			},
			"required": []string{"location"},
		},
	)
	
	return append(tools, weatherTool)
}
```

### Using QueryWithOptions for Advanced Features

The `QueryWithOptions` method provides access to advanced features like function calling:

```go
// Basic query (legacy method)
response, model, err := router.Query(ctx, messages, 0.7, "")

// Advanced query with options
options := gollmrouter.QueryOptions{
	Temperature: 0.7,
	ForceModel:  "gemini-2.0-flash", // Force specific model
	Tools:       []gollmrouter.Tool{...},
	ToolChoice:  "auto", // or "none" or specific tool name
}

result, err := router.QueryWithOptions(ctx, messages, options)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Content: %s\n", result.Content)
fmt.Printf("Model: %s\n", result.Model)
fmt.Printf("Function Calls: %+v\n", result.ToolCalls)
fmt.Printf("Finish Reason: %s\n", result.FinishReason)
```

## API Reference

### Core Types

#### Message
```go
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Files   []File `json:"files,omitempty"`
}
```

#### File
```go
type File struct {
	Type     string `json:"type"`      // "image", "document", etc.
	Data     []byte `json:"data"`      // file data
	MimeType string `json:"mime_type"` // MIME type
	Name     string `json:"name"`      // filename
}
```

#### Function Calling Types
```go
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolCallResult struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}
```

#### Query Options
```go
type QueryOptions struct {
	Temperature float64 `json:"temperature"`
	ForceModel  string  `json:"force_model,omitempty"`
	Tools       []Tool  `json:"tools,omitempty"`
	ToolChoice  string  `json:"tool_choice,omitempty"` // "auto", "none", or specific tool name
}
```

#### Query Result
```go
type QueryResult struct {
	Content      string     `json:"content"`
	Model        string     `json:"model"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"`
}
```

### Provider Configurations

#### GeminiConfig
```go
type GeminiConfig struct {
	APIKey       string
	Models       []string
	MaxDailyReqs int
}
```

#### OpenRouterConfig
```go
type OpenRouterConfig struct {
	APIKey       string
	Models       []string
	MaxDailyReqs int
	Referer      string
	XTitle       string
	Timeout      time.Duration
}
```

**Example:**
```go
cfg := gollmrouter.OpenRouterConfig{
    APIKey:       apiKey,
    Models:       []string{"mistralai/mistral-7b-instruct:free"},
    MaxDailyReqs: 1000,
    Referer:      "https://myapp.com",
    XTitle:       "My App",
    Timeout:      30 * time.Second,
}
```

#### FunctionCallingConfig
```go
type FunctionCallingConfig struct {
	APIKey       string
	URL          string
	Models       []string
	MaxDailyReqs int
	Timeout      time.Duration
	ToolExecutor ToolExecutor
}
```



### Helper Functions

#### File Attachments
```go
// Create file attachment from data
file := gollmrouter.NewFileAttachment("image", "image/jpeg", "photo.jpg", data)

// Create file attachment from file path
file, err := gollmrouter.NewFileAttachmentFromPath("path/to/image.jpg")

// Create message with image
message, err := gollmrouter.NewMessageWithImage("user", "What's in this image?", "path/to/image.jpg")
```

#### Tool Creation
```go
// Create a tool
tool := gollmrouter.NewTool(
	"get_weather",
	"Get weather information for a location",
	map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "City name or location",
			},
		},
		"required": []string{"location"},
	},
)

// Create a tool call
toolCall := gollmrouter.NewToolCall("call_123", "get_weather", map[string]interface{}{
	"location": "New York",
})

// Create a tool call result
result := gollmrouter.NewToolCallResult("call_123", "22°C, Sunny")
```

## How Fallback Works

The library provides **two levels of fallback** to maximize reliability:

### 1. Provider-Level Fallback (Model Fallback)
Each provider accepts a slice of models and will try them in order until one succeeds:
- If the first model fails (API error, rate limit, etc.), it automatically tries the next model
- This handles temporary issues with specific models or providers
- Example: `[]string{"gemini-2.0-flash", "gemini-1.5-flash"}` - if `gemini-2.0-flash` fails, it tries `gemini-1.5-flash`
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

func (p *MyCustomProvider) QueryWithOptions(ctx context.Context, messages []providers.Message, options providers.QueryOptions) (*providers.QueryResult, error) {
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
7. **Function Calling Support**: Providers can execute function calls and handle function calling workflows

## Supported Providers
- **Google Gemini**: Direct API integration with quota management, image support, and **full function calling** using the latest official SDK
- **OpenRouter**: OpenAI-compatible API gateway with access to multiple models and function calling support
- **Function Calling Provider**: Generic provider for any LLM API that supports function calling (OpenAI, Anthropic, etc.)

## Examples

See the `examples/` directory for complete working examples:
- `examples/mcp_example.go` - Comprehensive function calling examples with all providers

## Keywords
LLM, AI, Router, Fallback, Quota Management, Gemini, OpenRouter, OpenAI, Claude, API, Go, Golang, Library, File Attachments, Image Analysis, Vision Models, Function Calling, Tool Calls

## License
MIT

