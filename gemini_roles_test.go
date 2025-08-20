package gollmrouter_test

import (
	"testing"

	gollmrouter "github.com/FramnkRulez/go-llm-router"
)

func TestGeminiRoleValidation(t *testing.T) {
	// Test valid roles
	validRoles := []string{"user", "assistant", "system"}
	for _, role := range validRoles {
		message := gollmrouter.Message{Role: role, Content: "test"}
		err := gollmrouter.ValidateGeminiMessage(message)
		if err != nil {
			t.Errorf("Expected role '%s' to be valid, but got error: %v", role, err)
		}
	}

	// Test invalid roles
	invalidRoles := []string{"invalid", "bot", "admin", "moderator"}
	for _, role := range invalidRoles {
		message := gollmrouter.Message{Role: role, Content: "test"}
		err := gollmrouter.ValidateGeminiMessage(message)
		if err == nil {
			t.Errorf("Expected role '%s' to be invalid, but validation passed", role)
		}
	}
}

func TestGeminiMessageCreation(t *testing.T) {
	// Test user message creation
	userMessage := gollmrouter.NewGeminiUserMessage("Hello, how are you?")
	if userMessage.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", userMessage.Role)
	}
	if userMessage.Content != "Hello, how are you?" {
		t.Errorf("Expected content 'Hello, how are you?', got '%s'", userMessage.Content)
	}

	// Test model message creation
	modelMessage := gollmrouter.NewGeminiModelMessage("I'm doing well, thank you!")
	if modelMessage.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", modelMessage.Role)
	}
	if modelMessage.Content != "I'm doing well, thank you!" {
		t.Errorf("Expected content 'I'm doing well, thank you!', got '%s'", modelMessage.Content)
	}
	if len(modelMessage.Files) != 0 {
		t.Errorf("Expected no files in model message, got %d files", len(modelMessage.Files))
	}
}

func TestGeminiMessageWithFiles(t *testing.T) {
	// Create a file attachment
	fileAttachment := gollmrouter.NewFileAttachment("image", "image/jpeg", "test.jpg", []byte("fake image data"))

	// Test user message with files
	userMessage := gollmrouter.NewGeminiUserMessage("What's in this image?", fileAttachment)
	if userMessage.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", userMessage.Role)
	}
	if len(userMessage.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(userMessage.Files))
	}
	if userMessage.Files[0].Type != "image" {
		t.Errorf("Expected file type 'image', got '%s'", userMessage.Files[0].Type)
	}
}

func TestGeminiMessagesValidation(t *testing.T) {
	// Test valid messages
	validMessages := []gollmrouter.Message{
		gollmrouter.NewGeminiUserMessage("Hello"),
		gollmrouter.NewGeminiModelMessage("Hi there"),
		gollmrouter.NewGeminiUserMessage("How are you?"),
	}

	err := gollmrouter.ValidateGeminiMessages(validMessages)
	if err != nil {
		t.Errorf("Expected valid messages to pass validation, but got error: %v", err)
	}

	// Test invalid messages
	invalidMessages := []gollmrouter.Message{
		gollmrouter.NewGeminiUserMessage("Hello"),
		{Role: "invalid", Content: "This should fail"},
		gollmrouter.NewGeminiModelMessage("Hi there"),
	}

	err = gollmrouter.ValidateGeminiMessages(invalidMessages)
	if err == nil {
		t.Error("Expected invalid messages to fail validation, but validation passed")
	}
}

func TestGeminiRoleConversion(t *testing.T) {
	// Test that the internal conversion works correctly
	// This tests the internal convertRoleToGemini function indirectly

	// Create a Gemini provider to test role conversion
	geminiProvider, err := gollmrouter.NewGeminiProvider(gollmrouter.GeminiConfig{
		APIKey:       "test-api-key",
		Models:       []string{"gemini-2.0-flash"},
		MaxDailyReqs: 100,
	})
	if err != nil {
		t.Fatalf("Failed to create Gemini provider: %v", err)
	}
	defer geminiProvider.Close()

	// Test that the provider can handle valid roles
	validMessages := []gollmrouter.Message{
		gollmrouter.NewGeminiUserMessage("Test message"),
	}

	// This should not panic or error due to role validation
	// We're not making actual API calls, just testing the validation
	err = gollmrouter.ValidateGeminiMessages(validMessages)
	if err != nil {
		t.Errorf("Expected valid messages to pass validation: %v", err)
	}
}

func TestGeminiMessageWithImageCreation(t *testing.T) {
	// Test creating a user message with image (this will fail without a real file, but we can test the function signature)
	// We'll test with a non-existent file to ensure proper error handling
	_, err := gollmrouter.NewGeminiUserMessageWithImage("What's in this image?", "/non/existent/file.jpg")
	if err == nil {
		t.Error("Expected error when creating message with non-existent image file")
	}
}

func TestGeminiRoleConstants(t *testing.T) {
	// Test that the role constants are correctly defined
	// This tests the internal GeminiRole type indirectly through validation

	// Test that "user" is the default role for questions
	userMessage := gollmrouter.NewGeminiUserMessage("What is the weather?")
	if userMessage.Role != "user" {
		t.Errorf("Expected 'user' to be the default role for questions, got '%s'", userMessage.Role)
	}

	// Test that "assistant" is the role for model responses
	modelMessage := gollmrouter.NewGeminiModelMessage("The weather is sunny.")
	if modelMessage.Role != "assistant" {
		t.Errorf("Expected 'assistant' to be the role for model responses, got '%s'", modelMessage.Role)
	}
}

func TestGeminiMessageValidationEdgeCases(t *testing.T) {
	// Test empty content
	emptyMessage := gollmrouter.Message{Role: "user", Content: ""}
	err := gollmrouter.ValidateGeminiMessage(emptyMessage)
	if err != nil {
		t.Errorf("Expected empty content to be valid: %v", err)
	}

	// Test very long content
	longContent := string(make([]byte, 10000)) // 10KB of content
	longMessage := gollmrouter.Message{Role: "user", Content: longContent}
	err = gollmrouter.ValidateGeminiMessage(longMessage)
	if err != nil {
		t.Errorf("Expected long content to be valid: %v", err)
	}

	// Test with special characters
	specialMessage := gollmrouter.Message{Role: "user", Content: "Hello 世界! 🌍"}
	err = gollmrouter.ValidateGeminiMessage(specialMessage)
	if err != nil {
		t.Errorf("Expected special characters to be valid: %v", err)
	}
}
