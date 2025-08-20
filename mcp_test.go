package gollmrouter_test

import (
	"testing"

	"github.com/FramnkRulez/go-llm-router/ai"
	gollmrouter "github.com/FramnkRulez/go-llm-router"
)

func TestFunctionCallingProvider(t *testing.T) {
	// Test tool creation
	tool := gollmrouter.NewTool(
		"test_function",
		"A test function",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param": map[string]interface{}{
					"type":        "string",
					"description": "Test parameter",
				},
			},
			"required": []string{"param"},
		},
	)

	if tool.Function.Name != "test_function" {
		t.Errorf("Expected tool name 'test_function', got '%s'", tool.Function.Name)
	}

	// Test tool call creation
	toolCall := gollmrouter.NewToolCall("call_123", "test_function", map[string]interface{}{
		"param": "test_value",
	})

	if toolCall.ID != "call_123" {
		t.Errorf("Expected tool call ID 'call_123', got '%s'", toolCall.ID)
	}

	if toolCall.Function.Name != "test_function" {
		t.Errorf("Expected function name 'test_function', got '%s'", toolCall.Function.Name)
	}

	// Test tool call result creation
	result := gollmrouter.NewToolCallResult("call_123", "test_result")
	if result.ID != "call_123" {
		t.Errorf("Expected result ID 'call_123', got '%s'", result.ID)
	}

	if result.Content != "test_result" {
		t.Errorf("Expected result content 'test_result', got '%v'", result.Content)
	}

	// Test tool executor
	executor := ai.NewSimpleToolExecutor()
	tools := executor.GetAvailableTools()

	if len(tools) == 0 {
		t.Error("Expected tools to be available from SimpleToolExecutor")
	}

	// Test query options
	options := gollmrouter.QueryOptions{
		Temperature: 0.7,
		Tools:       tools,
		ToolChoice:  "auto",
	}

	if options.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", options.Temperature)
	}

	if options.ToolChoice != "auto" {
		t.Errorf("Expected tool choice 'auto', got '%s'", options.ToolChoice)
	}

	if len(options.Tools) == 0 {
		t.Error("Expected tools to be set in options")
	}
}

func TestFunctionCallingConfig(t *testing.T) {
	// Test that FunctionCallingConfig can be created
	config := gollmrouter.FunctionCallingConfig{
		APIKey:       "test-key",
		URL:          "https://api.openai.com/v1/chat/completions",
		Models:       []string{"gpt-4"},
		MaxDailyReqs: 100,
		ToolExecutor: ai.NewSimpleToolExecutor(),
	}

	if config.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", config.APIKey)
	}

	if config.URL != "https://api.openai.com/v1/chat/completions" {
		t.Errorf("Expected URL 'https://api.openai.com/v1/chat/completions', got '%s'", config.URL)
	}

	if len(config.Models) != 1 || config.Models[0] != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%v'", config.Models)
	}

	if config.MaxDailyReqs != 100 {
		t.Errorf("Expected max daily requests 100, got %d", config.MaxDailyReqs)
	}

	if config.ToolExecutor == nil {
		t.Error("Expected tool executor to be set")
	}
}

func TestToolExecutorInterface(t *testing.T) {
	executor := ai.NewSimpleToolExecutor()

	// Test that the executor implements the interface
	var _ gollmrouter.ToolExecutor = executor

	// Test getting available tools
	tools := executor.GetAvailableTools()
	if len(tools) == 0 {
		t.Error("Expected tools to be available")
	}

	// Test that tools have required fields
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
}

func TestQueryOptions(t *testing.T) {
	// Test creating query options
	options := gollmrouter.QueryOptions{
		Temperature: 0.8,
		ForceModel:  "gpt-4",
		ToolChoice:  "none",
	}

	if options.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8, got %f", options.Temperature)
	}

	if options.ForceModel != "gpt-4" {
		t.Errorf("Expected force model 'gpt-4', got '%s'", options.ForceModel)
	}

	if options.ToolChoice != "none" {
		t.Errorf("Expected tool choice 'none', got '%s'", options.ToolChoice)
	}

	// Test with tools
	tools := []gollmrouter.Tool{
		gollmrouter.NewTool("test", "test tool", map[string]interface{}{}),
	}
	options.Tools = tools

	if len(options.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(options.Tools))
	}
}
