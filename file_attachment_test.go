package gollmrouter_test

import (
	"testing"

	gollmrouter "github.com/FramnkRulez/go-llm-router"
)

func TestNewFileAttachment(t *testing.T) {
	data := []byte("test image data")
	file := gollmrouter.NewFileAttachment("image", "image/jpeg", "test.jpg", data)

	if file.Type != "image" {
		t.Errorf("Expected file type 'image', got '%s'", file.Type)
	}
	if file.MimeType != "image/jpeg" {
		t.Errorf("Expected MIME type 'image/jpeg', got '%s'", file.MimeType)
	}
	if file.Name != "test.jpg" {
		t.Errorf("Expected name 'test.jpg', got '%s'", file.Name)
	}
	if len(file.Data) != len(data) {
		t.Errorf("Expected data length %d, got %d", len(data), len(file.Data))
	}
}

func TestNewMessage(t *testing.T) {
	file := gollmrouter.NewFileAttachment("image", "image/jpeg", "test.jpg", []byte("data"))
	message := gollmrouter.NewMessage("user", "test content", file)

	if message.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", message.Role)
	}
	if message.Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", message.Content)
	}
	if len(message.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(message.Files))
	}
	if message.Files[0].Type != "image" {
		t.Errorf("Expected file type 'image', got '%s'", message.Files[0].Type)
	}
}

func TestNewMessageWithMultipleFiles(t *testing.T) {
	file1 := gollmrouter.NewFileAttachment("image", "image/jpeg", "test1.jpg", []byte("data1"))
	file2 := gollmrouter.NewFileAttachment("image", "image/png", "test2.png", []byte("data2"))
	message := gollmrouter.NewMessage("user", "test content", file1, file2)

	if len(message.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(message.Files))
	}
	if message.Files[0].Name != "test1.jpg" {
		t.Errorf("Expected first file name 'test1.jpg', got '%s'", message.Files[0].Name)
	}
	if message.Files[1].Name != "test2.png" {
		t.Errorf("Expected second file name 'test2.png', got '%s'", message.Files[1].Name)
	}
}

func TestNewMessageWithoutFiles(t *testing.T) {
	message := gollmrouter.NewMessage("user", "test content")

	if message.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", message.Role)
	}
	if message.Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", message.Content)
	}
	if len(message.Files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(message.Files))
	}
}
