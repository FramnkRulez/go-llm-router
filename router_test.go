package gollmrouter_test

import (
	"testing"

	gollmrouter "github.com/FramnkRulez/go-llm-router"
	"github.com/FramnkRulez/go-llm-router/provider"
)

func TestRouter_NoProviders(t *testing.T) {
	_, err := gollmrouter.NewRouter()
	if err == nil {
		t.Fatal("expected error when no providers are configured")
	}
}

func TestRouter_HasRemainingRequests_Empty(t *testing.T) {
	providers := []provider.Provider{}

	_, err := gollmrouter.NewRouter(providers...)
	if err == nil {
		t.Fatal("expected error when no providers are configured")
	}
}

// Add more tests with mock providers as needed
