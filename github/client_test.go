package github

import (
	"testing"
	"time"

	"github.com/google/go-github/v68/github"
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

func TestConvertRuns(t *testing.T) {
	// Helper to create pointers
	intPtr := func(i int64) *int64 { return &i }
	intValPtr := func(i int) *int { return &i }
	strPtr := func(s string) *string { return &s }

	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	ghTimestamp := &github.Timestamp{Time: createdAt}

	ghRuns := []*github.WorkflowRun{
		{
			ID:         intPtr(12345678901),
			RunNumber:  intValPtr(21),
			Name:       strPtr("CI"),
			Status:     strPtr("completed"),
			Conclusion: strPtr("success"),
			HeadBranch: strPtr("main"),
			Event:      strPtr("push"),
			Actor:      &github.User{Login: strPtr("testuser")},
			HTMLURL:    strPtr("https://github.com/owner/repo/actions/runs/12345678901"),
			CreatedAt:  ghTimestamp,
		},
		{
			ID:         intPtr(12345678902),
			RunNumber:  intValPtr(22),
			Name:       strPtr("Deploy"),
			Status:     strPtr("in_progress"),
			Conclusion: strPtr(""),
			HeadBranch: strPtr("feature-branch"),
			Event:      strPtr("pull_request"),
			Actor:      &github.User{Login: strPtr("anotheruser")},
			HTMLURL:    strPtr("https://github.com/owner/repo/actions/runs/12345678902"),
			CreatedAt:  ghTimestamp,
		},
	}

	runs := convertRuns(ghRuns)

	if len(runs) != 2 {
		t.Fatalf("convertRuns() returned %d runs, want 2", len(runs))
	}

	// Test first run
	r1 := runs[0]
	if r1.ID != 12345678901 {
		t.Errorf("Run[0].ID = %d, want 12345678901", r1.ID)
	}
	if r1.RunNumber != 21 {
		t.Errorf("Run[0].RunNumber = %d, want 21", r1.RunNumber)
	}
	if r1.Name != "CI" {
		t.Errorf("Run[0].Name = %q, want CI", r1.Name)
	}
	if r1.Status != "completed" {
		t.Errorf("Run[0].Status = %q, want completed", r1.Status)
	}
	if r1.Conclusion != "success" {
		t.Errorf("Run[0].Conclusion = %q, want success", r1.Conclusion)
	}
	if r1.Branch != "main" {
		t.Errorf("Run[0].Branch = %q, want main", r1.Branch)
	}
	if r1.Event != "push" {
		t.Errorf("Run[0].Event = %q, want push", r1.Event)
	}
	if r1.Actor != "testuser" {
		t.Errorf("Run[0].Actor = %q, want testuser", r1.Actor)
	}

	// Test second run
	r2 := runs[1]
	if r2.RunNumber != 22 {
		t.Errorf("Run[1].RunNumber = %d, want 22", r2.RunNumber)
	}
	if r2.Event != "pull_request" {
		t.Errorf("Run[1].Event = %q, want pull_request", r2.Event)
	}
	if r2.Branch != "feature-branch" {
		t.Errorf("Run[1].Branch = %q, want feature-branch", r2.Branch)
	}
}

func TestConvertRuns_EmptyInput(t *testing.T) {
	runs := convertRuns(nil)
	if len(runs) != 0 {
		t.Errorf("convertRuns(nil) returned %d runs, want 0", len(runs))
	}

	runs = convertRuns([]*github.WorkflowRun{})
	if len(runs) != 0 {
		t.Errorf("convertRuns([]) returned %d runs, want 0", len(runs))
	}
}
