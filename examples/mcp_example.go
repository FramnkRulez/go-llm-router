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

	// Create Gemini provider with function calling support (using the new official SDK)
	geminiProvider, err := gollmrouter.NewGeminiProvider(gollmrouter.GeminiConfig{
		APIKey:       "your-gemini-api-key",                            // Replace with your actual Gemini API key
		Models:       []string{"gemini-2.0-flash", "gemini-1.5-flash"}, // Latest Gemini models
		MaxDailyReqs: 100,
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini provider: %v", err)
	}

	// Create function calling provider for LLM APIs that support function calling
	// This works with OpenAI, Anthropic, or any other LLM API that supports the OpenAI function calling format
	functionCallingProvider, err := gollmrouter.NewFunctionCallingProvider(gollmrouter.FunctionCallingConfig{
		APIKey:       "your-openai-api-key",                        // Replace with your actual API key
		URL:          "https://api.openai.com/v1/chat/completions", // OpenAI API endpoint
		Models:       []string{"gpt-4", "gpt-3.5-turbo"},
		MaxDailyReqs: 100,
		Timeout:      30 * time.Second,
		ToolExecutor: toolExecutor, // Your local tool executor that runs functions
	})
	if err != nil {
		log.Fatalf("Failed to create function calling provider: %v", err)
	}

	// Create OpenRouter provider as fallback (also supports function calling)
	openRouterProvider, err := gollmrouter.NewOpenRouterProvider(gollmrouter.OpenRouterConfig{
		APIKey:       "your-openrouter-api-key",
		URL:          "https://openrouter.ai/api/v1/chat/completions",
		Models:       []string{"openai/gpt-4", "openai/gpt-3.5-turbo"},
		MaxDailyReqs: 100,
		Referer:      "your-app",
		XTitle:       "your-title",
		Timeout:      30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create OpenRouter provider: %v", err)
	}

	// Create router with all providers (Gemini, Function Calling, OpenRouter)
	router, err := gollmrouter.NewRouter(geminiProvider, functionCallingProvider, openRouterProvider)
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}
	defer router.Close()

	// Example 1: Basic query without function calls
	fmt.Println("=== Example 1: Basic Query ===")
	response, model, err := router.Query(context.Background(), []gollmrouter.Message{
		{Role: "user", Content: "Hello, who are you?"},
	}, 0.7, "")
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		fmt.Printf("Model: %s\nResponse: %s\n\n", model, response)
	}

	// Example 2: Query with function calls using QueryWithOptions
	fmt.Println("=== Example 2: Query with Function Calls ===")
	options := gollmrouter.QueryOptions{
		Temperature: 0.7,
		Tools:       toolExecutor.GetAvailableTools(),
		ToolChoice:  "auto",
	}

	result, err := router.QueryWithOptions(context.Background(), []gollmrouter.Message{
		{Role: "user", Content: "What is the current time and calculate 15 * 23?"},
	}, options)
	if err != nil {
		log.Printf("Query with options failed: %v", err)
	} else {
		fmt.Printf("Model: %s\nContent: %s\n", result.Model, result.Content)
		if len(result.ToolCalls) > 0 {
			fmt.Printf("Function Calls: %+v\n", result.ToolCalls)
		}
		fmt.Printf("Finish Reason: %s\n\n", result.FinishReason)
	}

	// Example 3: String operations with function calls
	fmt.Println("=== Example 3: String Operations ===")
	result2, err := router.QueryWithOptions(context.Background(), []gollmrouter.Message{
		{Role: "user", Content: "Convert 'Hello World' to uppercase and tell me its length."},
	}, options)
	if err != nil {
		log.Printf("String operations query failed: %v", err)
	} else {
		fmt.Printf("Model: %s\nContent: %s\n", result2.Model, result2.Content)
		if len(result2.ToolCalls) > 0 {
			fmt.Printf("Function Calls: %+v\n", result2.ToolCalls)
		}
		fmt.Printf("Finish Reason: %s\n\n", result2.FinishReason)
	}

	// Example 4: Mathematical calculations
	fmt.Println("=== Example 4: Mathematical Calculations ===")
	result3, err := router.QueryWithOptions(context.Background(), []gollmrouter.Message{
		{Role: "user", Content: "What is the square root of 144 and what is 25 + 37?"},
	}, options)
	if err != nil {
		log.Printf("Math query failed: %v", err)
	} else {
		fmt.Printf("Model: %s\nContent: %s\n", result3.Model, result3.Content)
		if len(result3.ToolCalls) > 0 {
			fmt.Printf("Function Calls: %+v\n", result3.ToolCalls)
		}
		fmt.Printf("Finish Reason: %s\n\n", result3.FinishReason)
	}

	// Example 5: Creating custom tools
	fmt.Println("=== Example 5: Custom Tools ===")
	customTools := []gollmrouter.Tool{
		gollmrouter.NewTool(
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
		),
	}

	customOptions := gollmrouter.QueryOptions{
		Temperature: 0.7,
		Tools:       customTools,
		ToolChoice:  "auto",
	}

	result4, err := router.QueryWithOptions(context.Background(), []gollmrouter.Message{
		{Role: "user", Content: "What's the weather like in New York?"},
	}, customOptions)
	if err != nil {
		log.Printf("Custom tools query failed: %v", err)
	} else {
		fmt.Printf("Model: %s\nContent: %s\n", result4.Model, result4.Content)
		if len(result4.ToolCalls) > 0 {
			fmt.Printf("Function Calls: %+v\n", result4.ToolCalls)
		}
		fmt.Printf("Finish Reason: %s\n\n", result4.FinishReason)
	}

	// Example 6: Testing Gemini-specific function calling
	fmt.Println("=== Example 6: Gemini Function Calling ===")
	geminiOptions := gollmrouter.QueryOptions{
		Temperature: 0.7,
		ForceModel:  "gemini-2.0-flash", // Force Gemini model
		Tools:       toolExecutor.GetAvailableTools(),
		ToolChoice:  "auto",
	}

	result5, err := router.QueryWithOptions(context.Background(), []gollmrouter.Message{
		{Role: "user", Content: "Calculate the area of a circle with radius 5 and tell me the current time."},
	}, geminiOptions)
	if err != nil {
		log.Printf("Gemini function calling query failed: %v", err)
	} else {
		fmt.Printf("Model: %s\nContent: %s\n", result5.Model, result5.Content)
		if len(result5.ToolCalls) > 0 {
			fmt.Printf("Function Calls: %+v\n", result5.ToolCalls)
		}
		fmt.Printf("Finish Reason: %s\n\n", result5.FinishReason)
	}
}

// Example of a custom tool executor for weather
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

		// Simulate weather data (in a real app, you'd call a weather API)
		weatherData := map[string]interface{}{
			"location":    location,
			"temperature": "22°C",
			"condition":   "Sunny",
			"humidity":    "65%",
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
