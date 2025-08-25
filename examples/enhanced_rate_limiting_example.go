package main

import (
	"context"
	"fmt"
	"log"
	"time"

	gollmrouter "github.com/FramnkRulez/go-llm-router"
	"github.com/FramnkRulez/go-llm-router/ai"
)

func enhancedRateLimitingExample() {
	// Create a tool executor with basic tools
	toolExecutor := ai.NewSimpleToolExecutor()

	fmt.Println("=== Enhanced Rate Limiting and Ranking Example ===")
	fmt.Println()

	// Create multiple Gemini providers with different rate limits and ranks
	// Higher rank = higher priority
	geminiProvider1, err := gollmrouter.NewGeminiProvider(gollmrouter.GeminiConfig{
		APIKey:               "your-gemini-api-key-1",
		Models:               []string{"gemini-2.0-flash", "gemini-1.5-flash"},
		MaxDailyReqs:         50,   // 50 requests per day
		MaxRequestsPerMinute: 10,   // 10 requests per minute
		MaxTokensPerMinute:   1000, // 1000 tokens per minute
		Rank:                 3,    // Highest priority
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini provider 1: %v", err)
	}

	geminiProvider2, err := gollmrouter.NewGeminiProvider(gollmrouter.GeminiConfig{
		APIKey:               "your-gemini-api-key-2",
		Models:               []string{"gemini-1.5-flash"},
		MaxDailyReqs:         100, // 100 requests per day
		MaxRequestsPerMinute: 5,   // 5 requests per minute
		MaxTokensPerMinute:   500, // 500 tokens per minute
		Rank:                 2,   // Medium priority
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini provider 2: %v", err)
	}

	// Create OpenRouter provider as fallback with lower rank
	openRouterProvider, err := gollmrouter.NewOpenRouterProvider(gollmrouter.OpenRouterConfig{
		APIKey:               "your-openrouter-api-key",
		Models:               []string{"openai/gpt-4", "openai/gpt-3.5-turbo"},
		MaxDailyReqs:         200,  // 200 requests per day
		MaxRequestsPerMinute: 20,   // 20 requests per minute
		MaxTokensPerMinute:   2000, // 2000 tokens per minute
		Rank:                 1,    // Lowest priority (fallback)
		Referer:              "your-app",
		XTitle:               "your-title",
		Timeout:              30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create OpenRouter provider: %v", err)
	}

	// Create function calling provider with medium rank
	functionCallingProvider, err := gollmrouter.NewFunctionCallingProvider(gollmrouter.FunctionCallingConfig{
		APIKey:               "your-openai-api-key",
		URL:                  "https://api.openai.com/v1/chat/completions",
		Models:               []string{"gpt-4", "gpt-3.5-turbo"},
		MaxDailyReqs:         75,  // 75 requests per day
		MaxRequestsPerMinute: 8,   // 8 requests per minute
		MaxTokensPerMinute:   800, // 800 tokens per minute
		Rank:                 2,   // Same as geminiProvider2
		Timeout:              30 * time.Second,
		ToolExecutor:         toolExecutor,
	})
	if err != nil {
		log.Fatalf("Failed to create function calling provider: %v", err)
	}

	// Create router with all providers
	// Providers will be automatically sorted by rank (highest first)
	router, err := gollmrouter.NewRouter(
		openRouterProvider,      // Rank 1 (lowest)
		functionCallingProvider, // Rank 2
		geminiProvider2,         // Rank 2
		geminiProvider1,         // Rank 3 (highest)
	)
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}
	defer router.Close()

	fmt.Println("Provider Priority Order (by rank):")
	fmt.Println("1. Gemini Provider 1 (Rank 3) - 50/day, 10/min, 1000 tokens/min")
	fmt.Println("2. Gemini Provider 2 (Rank 2) - 100/day, 5/min, 500 tokens/min")
	fmt.Println("3. Function Calling Provider (Rank 2) - 75/day, 8/min, 800 tokens/min")
	fmt.Println("4. OpenRouter Provider (Rank 1) - 200/day, 20/min, 2000 tokens/min")
	fmt.Println()

	// Example 1: Basic query - should use highest ranked provider with remaining quota
	fmt.Println("=== Example 1: Basic Query ===")
	response, model, err := router.Query(context.Background(), []gollmrouter.Message{
		gollmrouter.NewGeminiUserMessage("Hello, who are you?"),
	}, 0.7, "")
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		fmt.Printf("Model: %s\nResponse: %s\n\n", model, response)
	}

	// Example 2: Query with function calls - should prioritize providers that support function calling
	fmt.Println("=== Example 2: Query with Function Calls ===")
	options := gollmrouter.QueryOptions{
		Temperature: 0.7,
		Tools:       toolExecutor.GetAvailableTools(),
		ToolChoice:  "auto",
	}

	result, err := router.QueryWithOptions(context.Background(), []gollmrouter.Message{
		gollmrouter.NewGeminiUserMessage("What is the current time and calculate 15 * 23?"),
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

	// Example 3: Test rate limiting by making multiple requests
	fmt.Println("=== Example 3: Rate Limiting Test ===")
	fmt.Println("Making 5 quick requests to test rate limiting...")

	for i := 1; i <= 5; i++ {
		fmt.Printf("Request %d: ", i)
		_, model, err := router.Query(context.Background(), []gollmrouter.Message{
			gollmrouter.NewGeminiUserMessage(fmt.Sprintf("Quick test request %d", i)),
		}, 0.7, "")
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
		} else {
			fmt.Printf("Success - Model: %s\n", model)
		}
		time.Sleep(100 * time.Millisecond) // Small delay between requests
	}
	fmt.Println()

	// Example 4: Test empty response handling
	fmt.Println("=== Example 4: Empty Response Handling ===")
	fmt.Println("Note: Empty responses are now treated as errors and will cause fallback to next provider")

	// This would normally trigger an empty response scenario
	// In a real scenario, some models might return empty responses to indicate failure
	_, err = router.QueryWithOptions(context.Background(), []gollmrouter.Message{
		gollmrouter.NewGeminiUserMessage("Generate a very short response"),
	}, gollmrouter.QueryOptions{
		Temperature: 0.1,
		ForceModel:  "gemini-1.5-flash", // Force a specific model
	})
	if err != nil {
		fmt.Printf("Query failed (as expected if empty response): %v\n", err)
	} else {
		fmt.Printf("Query succeeded\n")
	}
	fmt.Println()

	// Example 5: Check remaining capacity
	fmt.Println("=== Example 5: Remaining Capacity Check ===")
	if router.HasRemainingRequests(context.Background()) {
		fmt.Println("✓ Router has remaining capacity across all providers")
	} else {
		fmt.Println("✗ Router has no remaining capacity")
	}
	fmt.Println()

	// Example 6: Demonstrate provider ranking in action
	fmt.Println("=== Example 6: Provider Ranking Demonstration ===")
	fmt.Println("The router automatically tries providers in order of rank:")
	fmt.Println("- Higher ranked providers are tried first")
	fmt.Println("- If a provider hits rate limits, it moves to the next")
	fmt.Println("- Empty responses are treated as errors and trigger fallback")
	fmt.Println("- All rate limits (daily, per-minute, tokens) are checked")
	fmt.Println()

	// Example 7: Simulate hitting rate limits
	fmt.Println("=== Example 7: Rate Limit Simulation ===")
	fmt.Println("In a real scenario, when providers hit their rate limits:")
	fmt.Println("- Daily limits: Reset every 24 hours")
	fmt.Println("- Per-minute limits: Reset every minute")
	fmt.Println("- Token limits: Reset every minute")
	fmt.Println("- Router automatically falls back to next available provider")
	fmt.Println()

	fmt.Println("=== Configuration Tips ===")
	fmt.Println("1. Set Rank based on your preference:")
	fmt.Println("   - Rank 3: Premium/fastest providers")
	fmt.Println("   - Rank 2: Standard providers")
	fmt.Println("   - Rank 1: Fallback providers")
	fmt.Println()
	fmt.Println("2. Configure rate limits based on your API quotas:")
	fmt.Println("   - MaxDailyReqs: Set to your daily API limit")
	fmt.Println("   - MaxRequestsPerMinute: Set to your per-minute request limit")
	fmt.Println("   - MaxTokensPerMinute: Set to your per-minute token limit")
	fmt.Println()
	fmt.Println("3. Use 0 for unlimited (no limit enforced)")
	fmt.Println("4. Empty responses are automatically treated as errors")
	fmt.Println("5. Providers are automatically sorted by rank on router creation")
}
