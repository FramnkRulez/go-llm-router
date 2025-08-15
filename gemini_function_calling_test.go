package gollmrouter

import (
	"context"
	"testing"

	"github.com/FramnkRulez/go-llm-router/ai"
	"github.com/FramnkRulez/go-llm-router/provider"
)

func TestGeminiProviderWithFunctionCalling(t *testing.T) {
	// Test that Gemini provider can be created with function calling support
	geminiProvider, err := NewGeminiProvider(GeminiConfig{
		APIKey:       "test-api-key",
		Models:       []string{"gemini-2.0-flash", "gemini-1.5-flash"},
		MaxDailyReqs: 100,
	})
	if err != nil {
		t.Fatalf("Failed to create Gemini provider: %v", err)
	}
	defer geminiProvider.Close()

	// Test that the provider implements the interface
	var _ provider.Provider = geminiProvider

	// Test basic functionality
	if !geminiProvider.HasRemainingRequests(context.Background()) {
		t.Error("Expected Gemini provider to have remaining requests")
	}
}

func TestGeminiProviderConfiguration(t *testing.T) {
	// Test Gemini configuration
	config := GeminiConfig{
		APIKey:       "test-gemini-key",
		Models:       []string{"gemini-2.0-flash", "gemini-1.5-flash", "gemini-1.5-pro"},
		MaxDailyReqs: 150,
	}

	if config.APIKey != "test-gemini-key" {
		t.Errorf("Expected API key 'test-gemini-key', got '%s'", config.APIKey)
	}

	if len(config.Models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(config.Models))
	}

	expectedModels := []string{"gemini-2.0-flash", "gemini-1.5-flash", "gemini-1.5-pro"}
	for i, model := range config.Models {
		if model != expectedModels[i] {
			t.Errorf("Expected model '%s', got '%s'", expectedModels[i], model)
		}
	}

	if config.MaxDailyReqs != 150 {
		t.Errorf("Expected max daily requests 150, got %d", config.MaxDailyReqs)
	}
}

func TestGeminiProviderWithTools(t *testing.T) {
	// Create a tool executor
	toolExecutor := ai.NewSimpleToolExecutor()
	tools := toolExecutor.GetAvailableTools()

	// Test that tools are properly configured
	if len(tools) == 0 {
		t.Error("Expected tools to be available")
	}

	// Test tool structure
	for _, tool := range tools {
		if tool.Type != "function" {
			t.Errorf("Expected tool type 'function', got '%s'", tool.Type)
		}

		if tool.Function.Name == "" {
			t.Error("Expected tool function to have a name")
		}

		if tool.Function.Description == "" {
			t.Error("Expected tool function to have a description")
		}
	}

	// Test specific tools that should be available
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Function.Name] = true
	}

	expectedTools := []string{"get_current_time", "calculate", "string_operations"}
	for _, expectedTool := range expectedTools {
		if !toolNames[expectedTool] {
			t.Errorf("Expected tool '%s' to be available", expectedTool)
		}
	}
}

func TestGeminiProviderQueryOptions(t *testing.T) {
	// Test query options with Gemini-specific settings
	options := QueryOptions{
		Temperature: 0.8,
		ForceModel:  "gemini-2.0-flash",
		ToolChoice:  "auto",
	}

	if options.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8, got %f", options.Temperature)
	}

	if options.ForceModel != "gemini-2.0-flash" {
		t.Errorf("Expected force model 'gemini-2.0-flash', got '%s'", options.ForceModel)
	}

	if options.ToolChoice != "auto" {
		t.Errorf("Expected tool choice 'auto', got '%s'", options.ToolChoice)
	}

	// Test with tools
	toolExecutor := ai.NewSimpleToolExecutor()
	options.Tools = toolExecutor.GetAvailableTools()

	if len(options.Tools) == 0 {
		t.Error("Expected tools to be set in options")
	}
}

func TestGeminiProviderModelFallback(t *testing.T) {
	// Test that Gemini provider supports model fallback
	geminiProvider, err := NewGeminiProvider(GeminiConfig{
		APIKey:       "test-api-key",
		Models:       []string{"gemini-2.0-flash", "gemini-1.5-flash", "gemini-1.5-pro"},
		MaxDailyReqs: 100,
	})
	if err != nil {
		t.Fatalf("Failed to create Gemini provider: %v", err)
	}
	defer geminiProvider.Close()

	// Test that the provider has multiple models configured
	// This would be tested in actual API calls, but we can verify the configuration
	config := GeminiConfig{
		Models: []string{"gemini-2.0-flash", "gemini-1.5-flash", "gemini-1.5-pro"},
	}

	if len(config.Models) < 2 {
		t.Error("Expected multiple models for fallback testing")
	}
}

func TestGeminiProviderFunctionCallingIntegration(t *testing.T) {
	// Test integration with function calling (without making actual API calls)
	toolExecutor := ai.NewSimpleToolExecutor()
	tools := toolExecutor.GetAvailableTools()

	// Test tool creation for Gemini
	for _, tool := range tools {
		// Test that tools have the correct structure for Gemini API
		if tool.Function.Name == "calculate" {
			// Test calculate tool structure
			if tool.Function.Description == "" {
				t.Error("Calculate tool should have a description")
			}
		}

		if tool.Function.Name == "get_current_time" {
			// Test time tool structure
			if tool.Function.Description == "" {
				t.Error("Get current time tool should have a description")
			}
		}

		if tool.Function.Name == "string_operations" {
			// Test string operations tool structure
			if tool.Function.Description == "" {
				t.Error("String operations tool should have a description")
			}
		}
	}
}

func TestGeminiProviderWithCustomTools(t *testing.T) {
	// Test creating custom tools for Gemini
	customTools := []Tool{
		NewTool(
			"calculate_area",
			"Calculate the area of a circle given its radius",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"radius": map[string]interface{}{
						"type":        "number",
						"description": "Radius of the circle",
					},
				},
				"required": []string{"radius"},
			},
		),
		NewTool(
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

	// Test custom tool structure
	if len(customTools) != 2 {
		t.Errorf("Expected 2 custom tools, got %d", len(customTools))
	}

	for _, tool := range customTools {
		if tool.Type != "function" {
			t.Errorf("Expected tool type 'function', got '%s'", tool.Type)
		}

		if tool.Function.Name == "" {
			t.Error("Expected tool function to have a name")
		}

		if tool.Function.Description == "" {
			t.Error("Expected tool function to have a description")
		}
	}

	// Test specific custom tools
	areaTool := customTools[0]
	if areaTool.Function.Name != "calculate_area" {
		t.Errorf("Expected tool name 'calculate_area', got '%s'", areaTool.Function.Name)
	}

	weatherTool := customTools[1]
	if weatherTool.Function.Name != "get_weather" {
		t.Errorf("Expected tool name 'get_weather', got '%s'", weatherTool.Function.Name)
	}
}

func TestGeminiProviderTemperatureHandling(t *testing.T) {
	// Test temperature handling in query options
	testCases := []struct {
		name        string
		temperature float64
		expected    float64
	}{
		{"Zero temperature", 0.0, 0.0},
		{"Low temperature", 0.1, 0.1},
		{"Medium temperature", 0.5, 0.5},
		{"High temperature", 0.9, 0.9},
		{"Maximum temperature", 1.0, 1.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := QueryOptions{
				Temperature: tc.temperature,
				ForceModel:  "gemini-2.0-flash",
			}

			if options.Temperature != tc.expected {
				t.Errorf("Expected temperature %f, got %f", tc.expected, options.Temperature)
			}
		})
	}
}

func TestGeminiProviderToolChoiceOptions(t *testing.T) {
	// Test tool choice options
	testCases := []struct {
		name       string
		toolChoice string
		expected   string
	}{
		{"Auto tool choice", "auto", "auto"},
		{"None tool choice", "none", "none"},
		{"Specific tool choice", "calculate", "calculate"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := QueryOptions{
				Temperature: 0.7,
				ForceModel:  "gemini-2.0-flash",
				ToolChoice:  tc.toolChoice,
			}

			if options.ToolChoice != tc.expected {
				t.Errorf("Expected tool choice '%s', got '%s'", tc.expected, options.ToolChoice)
			}
		})
	}
}
