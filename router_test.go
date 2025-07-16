package gollmrouter

import (
	"context"
	"testing"
)

func TestRouter_NoProviders(t *testing.T) {
	_, err := NewRouter()
	if err == nil {
		t.Fatal("expected error when no providers are configured")
	}
}

func TestRouter_HasRemainingRequests_Empty(t *testing.T) {
	r := &Router{providers: nil}
	if r.HasRemainingRequests(context.Background()) {
		t.Fatal("expected false when no providers")
	}
}

// Add more tests with mock providers as needed
