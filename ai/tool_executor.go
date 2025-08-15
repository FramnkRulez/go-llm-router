package ai

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/FramnkRulez/go-llm-router/provider"
)

// SimpleToolExecutor implements the ToolExecutor interface with basic tools
type SimpleToolExecutor struct {
	tools []provider.Tool
}

// NewSimpleToolExecutor creates a new simple tool executor with basic tools
func NewSimpleToolExecutor() *SimpleToolExecutor {
	executor := &SimpleToolExecutor{}
	executor.tools = []provider.Tool{
		{
			Type: "function",
			Function: provider.ToolFunction{
				Name:        "get_current_time",
				Description: "Get the current date and time",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"format": map[string]interface{}{
							"type":        "string",
							"description": "Time format (e.g., 'RFC3339', 'Unix')",
							"enum":        []string{"RFC3339", "Unix"},
						},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: "function",
			Function: provider.ToolFunction{
				Name:        "calculate",
				Description: "Perform mathematical calculations",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"expression": map[string]interface{}{
							"type":        "string",
							"description": "Mathematical expression to evaluate (e.g., '2 + 2', 'sqrt(16)')",
						},
					},
					"required": []string{"expression"},
				},
			},
		},
		{
			Type: "function",
			Function: provider.ToolFunction{
				Name:        "string_operations",
				Description: "Perform string operations like length, uppercase, lowercase",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"text": map[string]interface{}{
							"type":        "string",
							"description": "Text to operate on",
						},
						"operation": map[string]interface{}{
							"type":        "string",
							"description": "Operation to perform",
							"enum":        []string{"length", "uppercase", "lowercase", "reverse"},
						},
					},
					"required": []string{"text", "operation"},
				},
			},
		},
	}
	return executor
}

// ExecuteTool executes a tool call and returns the result
func (e *SimpleToolExecutor) ExecuteTool(ctx context.Context, toolCall provider.ToolCall) (*provider.ToolCallResult, error) {
	switch toolCall.Function.Name {
	case "get_current_time":
		return e.executeGetCurrentTime(toolCall)
	case "calculate":
		return e.executeCalculate(toolCall)
	case "string_operations":
		return e.executeStringOperations(toolCall)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}
}

// GetAvailableTools returns the list of available tools
func (e *SimpleToolExecutor) GetAvailableTools() []provider.Tool {
	return e.tools
}

// executeGetCurrentTime handles the get_current_time tool
func (e *SimpleToolExecutor) executeGetCurrentTime(toolCall provider.ToolCall) (*provider.ToolCallResult, error) {
	format := "RFC3339"
	if formatVal, ok := toolCall.Function.Arguments["format"]; ok {
		if formatStr, ok := formatVal.(string); ok {
			format = formatStr
		}
	}

	var result string
	switch format {
	case "RFC3339":
		result = time.Now().Format(time.RFC3339)
	case "Unix":
		result = strconv.FormatInt(time.Now().Unix(), 10)
	default:
		result = time.Now().Format(time.RFC3339)
	}

	return &provider.ToolCallResult{
		ID:      toolCall.ID,
		Type:    "function",
		Content: result,
	}, nil
}

// executeCalculate handles the calculate tool
func (e *SimpleToolExecutor) executeCalculate(toolCall provider.ToolCall) (*provider.ToolCallResult, error) {
	expression, ok := toolCall.Function.Arguments["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("expression argument is required")
	}

	// Simple expression evaluator - in a real implementation, you'd use a proper math parser
	result, err := e.evaluateExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return &provider.ToolCallResult{
		ID:      toolCall.ID,
		Type:    "function",
		Content: result,
	}, nil
}

// executeStringOperations handles the string_operations tool
func (e *SimpleToolExecutor) executeStringOperations(toolCall provider.ToolCall) (*provider.ToolCallResult, error) {
	text, ok := toolCall.Function.Arguments["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text argument is required")
	}

	operation, ok := toolCall.Function.Arguments["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation argument is required")
	}

	var result string
	switch operation {
	case "length":
		result = strconv.Itoa(len(text))
	case "uppercase":
		result = strings.ToUpper(text)
	case "lowercase":
		result = strings.ToLower(text)
	case "reverse":
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = string(runes)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	return &provider.ToolCallResult{
		ID:      toolCall.ID,
		Type:    "function",
		Content: result,
	}, nil
}

// evaluateExpression evaluates simple mathematical expressions
func (e *SimpleToolExecutor) evaluateExpression(expr string) (string, error) {
	expr = strings.TrimSpace(expr)

	// Handle sqrt function
	if strings.HasPrefix(expr, "sqrt(") && strings.HasSuffix(expr, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(expr, "sqrt("), ")")
		val, err := strconv.ParseFloat(inner, 64)
		if err != nil {
			return "", fmt.Errorf("invalid number in sqrt: %s", inner)
		}
		if val < 0 {
			return "", fmt.Errorf("cannot take square root of negative number")
		}
		return fmt.Sprintf("%.2f", math.Sqrt(val)), nil
	}

	// Handle basic arithmetic
	parts := strings.Fields(expr)
	if len(parts) != 3 {
		return "", fmt.Errorf("expression must be in format 'number operator number'")
	}

	a, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return "", fmt.Errorf("invalid first number: %s", parts[0])
	}

	b, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return "", fmt.Errorf("invalid second number: %s", parts[2])
	}

	var result float64
	switch parts[1] {
	case "+":
		result = a + b
	case "-":
		result = a - b
	case "*":
		result = a * b
	case "/":
		if b == 0 {
			return "", fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return "", fmt.Errorf("unknown operator: %s", parts[1])
	}

	return fmt.Sprintf("%.2f", result), nil
}
