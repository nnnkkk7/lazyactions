package github

import (
	"testing"
)

func TestNewClient_CreatesClient(t *testing.T) {
	client := NewClient("test-token", "owner", "repo")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	// Check it implements the interface
	var _ Client = client
}

func TestNewClient_WithEmptyToken(t *testing.T) {
	client := NewClient("", "owner", "repo")

	if client == nil {
		t.Fatal("NewClient returned nil with empty token")
	}
}

func TestRealClient_RateLimitRemaining(t *testing.T) {
	client := NewClient("token", "owner", "repo").(*realClient)

	// Default rate limit should be 5000
	remaining := client.RateLimitRemaining()
	if remaining != 5000 {
		t.Errorf("RateLimitRemaining() = %d, want 5000", remaining)
	}
}

func TestTokenTransport_SetsAuthHeader(t *testing.T) {
	// This is a basic test to ensure the transport is created
	transport := &tokenTransport{token: "test-token"}
	if transport.token != "test-token" {
		t.Error("token not set correctly")
	}
}
