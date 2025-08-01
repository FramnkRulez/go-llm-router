package gollmrouter

import (
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FramnkRulez/go-llm-router/internal/httpclient"
	"github.com/FramnkRulez/go-llm-router/internal/providers"
	"github.com/FramnkRulez/go-llm-router/provider"
)

// FileAttachment represents a file attachment with metadata and data
type FileAttachment = provider.File

// Message represents a chat message with role, content, and optional file attachments
type Message = provider.Message

// GeminiConfig holds configuration for creating a Gemini provider
type GeminiConfig struct {
	APIKey       string
	Models       []string
	MaxDailyReqs int
}

// OpenRouterConfig holds configuration for creating an OpenRouter provider
type OpenRouterConfig struct {
	APIKey       string
	URL          string
	Models       []string
	MaxDailyReqs int
	Referer      string
	XTitle       string
	Timeout      time.Duration
}

// NewGeminiProvider creates a new Gemini provider with the given configuration
func NewGeminiProvider(config GeminiConfig) (provider.Provider, error) {
	return providers.NewGeminiProvider(
		config.APIKey,
		config.Models,
		config.MaxDailyReqs,
	)
}

// NewOpenRouterProvider creates a new OpenRouter provider with the given configuration
func NewOpenRouterProvider(config OpenRouterConfig) (provider.Provider, error) {
	httpClient := httpclient.New("go-llm-router/1.0")

	return providers.NewOpenRouterProvider(
		config.APIKey,
		config.URL,
		config.Timeout,
		config.Models,
		config.Referer,
		config.XTitle,
		httpClient,
		config.MaxDailyReqs,
	)
}

// NewFileAttachment creates a new file attachment from file data
func NewFileAttachment(fileType, mimeType, name string, data []byte) FileAttachment {
	return FileAttachment{
		Type:     fileType,
		Data:     data,
		MimeType: mimeType,
		Name:     name,
	}
}

// NewFileAttachmentFromPath creates a file attachment from a file path
func NewFileAttachmentFromPath(filePath string) (FileAttachment, error) {
	// Read file data
	file, err := os.Open(filePath)
	if err != nil {
		return FileAttachment{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return FileAttachment{}, err
	}

	// Determine file type and MIME type
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Determine file type based on MIME type
	fileType := "document"
	if strings.HasPrefix(mimeType, "image/") {
		fileType = "image"
	}

	return FileAttachment{
		Type:     fileType,
		Data:     data,
		MimeType: mimeType,
		Name:     filepath.Base(filePath),
	}, nil
}

// NewMessage creates a new message with optional file attachments
func NewMessage(role, content string, files ...FileAttachment) Message {
	return Message{
		Role:    role,
		Content: content,
		Files:   files,
	}
}

// NewMessageWithImage creates a message with an image attachment
func NewMessageWithImage(role, content, imagePath string) (Message, error) {
	imageFile, err := NewFileAttachmentFromPath(imagePath)
	if err != nil {
		return Message{}, err
	}
	return NewMessage(role, content, imageFile), nil
}
