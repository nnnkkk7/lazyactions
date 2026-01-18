package github

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMockClient_ImplementsInterface(t *testing.T) {
	// Compile-time check that MockClient implements Client interface
	var _ Client = (*MockClient)(nil)
}

func TestMockClient_ListWorkflows(t *testing.T) {
	workflows := []Workflow{
		{ID: 1, Name: "CI", Path: ".github/workflows/ci.yml", State: "active"},
		{ID: 2, Name: "Deploy", Path: ".github/workflows/deploy.yml", State: "active"},
	}

	mock := &MockClient{Workflows: workflows}
	repo := Repository{Owner: "owner", Name: "repo"}

	got, err := mock.ListWorkflows(context.Background(), repo)
	if err != nil {
		t.Errorf("ListWorkflows() error = %v, want nil", err)
	}
	if len(got) != len(workflows) {
		t.Errorf("ListWorkflows() returned %d workflows, want %d", len(got), len(workflows))
	}

	// Verify call was recorded
	if mock.CallCount("ListWorkflows") != 1 {
		t.Errorf("CallCount(ListWorkflows) = %d, want 1", mock.CallCount("ListWorkflows"))
	}
}

func TestMockClient_ListWorkflows_Error(t *testing.T) {
	expectedErr := errors.New("API error")
	mock := &MockClient{Err: expectedErr}
	repo := Repository{Owner: "owner", Name: "repo"}

	_, err := mock.ListWorkflows(context.Background(), repo)
	if err != expectedErr {
		t.Errorf("ListWorkflows() error = %v, want %v", err, expectedErr)
	}
}

func TestMockClient_ListRuns(t *testing.T) {
	runs := []Run{
		{ID: 100, Name: "CI", Status: "completed", Conclusion: "success"},
		{ID: 101, Name: "CI", Status: "in_progress"},
	}

	mock := &MockClient{Runs: runs}
	repo := Repository{Owner: "owner", Name: "repo"}
	opts := &ListRunsOpts{Branch: "main"}

	got, err := mock.ListRuns(context.Background(), repo, opts)
	if err != nil {
		t.Errorf("ListRuns() error = %v, want nil", err)
	}
	if len(got) != len(runs) {
		t.Errorf("ListRuns() returned %d runs, want %d", len(got), len(runs))
	}

	// Verify call was recorded with args
	calls := mock.Calls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "ListRuns" {
		t.Errorf("Call method = %s, want ListRuns", calls[0].Method)
	}
}

func TestMockClient_ListJobs(t *testing.T) {
	jobs := []Job{
		{ID: 200, Name: "build", Status: "completed", Conclusion: "success"},
		{ID: 201, Name: "test", Status: "completed", Conclusion: "failure"},
	}

	mock := &MockClient{Jobs: jobs}
	repo := Repository{Owner: "owner", Name: "repo"}

	got, err := mock.ListJobs(context.Background(), repo, 100)
	if err != nil {
		t.Errorf("ListJobs() error = %v, want nil", err)
	}
	if len(got) != len(jobs) {
		t.Errorf("ListJobs() returned %d jobs, want %d", len(got), len(jobs))
	}

	if mock.CallCount("ListJobs") != 1 {
		t.Errorf("CallCount(ListJobs) = %d, want 1", mock.CallCount("ListJobs"))
	}
}

func TestMockClient_GetJobLogs(t *testing.T) {
	logs := "Running tests...\nAll tests passed!"

	mock := &MockClient{Logs: logs}
	repo := Repository{Owner: "owner", Name: "repo"}

	got, err := mock.GetJobLogs(context.Background(), repo, 200)
	if err != nil {
		t.Errorf("GetJobLogs() error = %v, want nil", err)
	}
	if got != logs {
		t.Errorf("GetJobLogs() = %q, want %q", got, logs)
	}

	if mock.CallCount("GetJobLogs") != 1 {
		t.Errorf("CallCount(GetJobLogs) = %d, want 1", mock.CallCount("GetJobLogs"))
	}
}

func TestMockClient_CancelRun(t *testing.T) {
	mock := &MockClient{}
	repo := Repository{Owner: "owner", Name: "repo"}

	err := mock.CancelRun(context.Background(), repo, 100)
	if err != nil {
		t.Errorf("CancelRun() error = %v, want nil", err)
	}

	if mock.CallCount("CancelRun") != 1 {
		t.Errorf("CallCount(CancelRun) = %d, want 1", mock.CallCount("CancelRun"))
	}
}

func TestMockClient_CancelRun_Error(t *testing.T) {
	expectedErr := errors.New("cannot cancel")
	mock := &MockClient{Err: expectedErr}
	repo := Repository{Owner: "owner", Name: "repo"}

	err := mock.CancelRun(context.Background(), repo, 100)
	if err != expectedErr {
		t.Errorf("CancelRun() error = %v, want %v", err, expectedErr)
	}
}

func TestMockClient_RerunWorkflow(t *testing.T) {
	mock := &MockClient{}
	repo := Repository{Owner: "owner", Name: "repo"}

	err := mock.RerunWorkflow(context.Background(), repo, 100)
	if err != nil {
		t.Errorf("RerunWorkflow() error = %v, want nil", err)
	}

	if mock.CallCount("RerunWorkflow") != 1 {
		t.Errorf("CallCount(RerunWorkflow) = %d, want 1", mock.CallCount("RerunWorkflow"))
	}
}

func TestMockClient_RerunFailedJobs(t *testing.T) {
	mock := &MockClient{}
	repo := Repository{Owner: "owner", Name: "repo"}

	err := mock.RerunFailedJobs(context.Background(), repo, 100)
	if err != nil {
		t.Errorf("RerunFailedJobs() error = %v, want nil", err)
	}

	if mock.CallCount("RerunFailedJobs") != 1 {
		t.Errorf("CallCount(RerunFailedJobs) = %d, want 1", mock.CallCount("RerunFailedJobs"))
	}
}

func TestMockClient_TriggerWorkflow(t *testing.T) {
	mock := &MockClient{}
	repo := Repository{Owner: "owner", Name: "repo"}
	inputs := map[string]interface{}{"environment": "production"}

	err := mock.TriggerWorkflow(context.Background(), repo, "ci.yml", "main", inputs)
	if err != nil {
		t.Errorf("TriggerWorkflow() error = %v, want nil", err)
	}

	if mock.CallCount("TriggerWorkflow") != 1 {
		t.Errorf("CallCount(TriggerWorkflow) = %d, want 1", mock.CallCount("TriggerWorkflow"))
	}

	// Verify args were recorded
	calls := mock.Calls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}
	if len(calls[0].Args) != 4 {
		t.Errorf("Expected 4 args, got %d", len(calls[0].Args))
	}
}

func TestMockClient_RateLimitRemaining(t *testing.T) {
	mock := &MockClient{}

	got := mock.RateLimitRemaining()
	if got != 5000 {
		t.Errorf("RateLimitRemaining() = %d, want 5000 (default)", got)
	}

	if mock.CallCount("RateLimitRemaining") != 1 {
		t.Errorf("CallCount(RateLimitRemaining) = %d, want 1", mock.CallCount("RateLimitRemaining"))
	}
}

func TestMockClient_RateLimitRemaining_Custom(t *testing.T) {
	mock := &MockClient{RateLimit: 100}

	got := mock.RateLimitRemaining()
	if got != 100 {
		t.Errorf("RateLimitRemaining() = %d, want 100", got)
	}
}

func TestMockClient_RecordCall(t *testing.T) {
	mock := &MockClient{}

	mock.RecordCall("TestMethod", "arg1", 123, true)

	calls := mock.Calls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	call := calls[0]
	if call.Method != "TestMethod" {
		t.Errorf("Method = %s, want TestMethod", call.Method)
	}
	if len(call.Args) != 3 {
		t.Errorf("Args length = %d, want 3", len(call.Args))
	}
	if call.Args[0] != "arg1" {
		t.Errorf("Args[0] = %v, want arg1", call.Args[0])
	}
	if call.Args[1] != 123 {
		t.Errorf("Args[1] = %v, want 123", call.Args[1])
	}
	if call.Args[2] != true {
		t.Errorf("Args[2] = %v, want true", call.Args[2])
	}
	if call.Time.IsZero() {
		t.Error("Call.Time should not be zero")
	}
}

func TestMockClient_CallCount(t *testing.T) {
	mock := &MockClient{}
	repo := Repository{Owner: "owner", Name: "repo"}

	// Make multiple calls
	mock.ListWorkflows(context.Background(), repo)
	mock.ListWorkflows(context.Background(), repo)
	mock.ListRuns(context.Background(), repo, nil)
	mock.ListWorkflows(context.Background(), repo)

	if mock.CallCount("ListWorkflows") != 3 {
		t.Errorf("CallCount(ListWorkflows) = %d, want 3", mock.CallCount("ListWorkflows"))
	}
	if mock.CallCount("ListRuns") != 1 {
		t.Errorf("CallCount(ListRuns) = %d, want 1", mock.CallCount("ListRuns"))
	}
	if mock.CallCount("NonExistent") != 0 {
		t.Errorf("CallCount(NonExistent) = %d, want 0", mock.CallCount("NonExistent"))
	}
}

func TestMockClient_Reset(t *testing.T) {
	mock := &MockClient{}
	repo := Repository{Owner: "owner", Name: "repo"}

	// Make some calls
	mock.ListWorkflows(context.Background(), repo)
	mock.ListRuns(context.Background(), repo, nil)

	if mock.CallCount("ListWorkflows") != 1 {
		t.Fatal("Expected 1 call before reset")
	}

	// Reset
	mock.Reset()

	if mock.CallCount("ListWorkflows") != 0 {
		t.Errorf("CallCount(ListWorkflows) after Reset() = %d, want 0", mock.CallCount("ListWorkflows"))
	}
	if len(mock.Calls()) != 0 {
		t.Errorf("Calls() after Reset() = %d, want 0", len(mock.Calls()))
	}
}

func TestMockClient_Calls_ReturnsCopy(t *testing.T) {
	mock := &MockClient{}
	repo := Repository{Owner: "owner", Name: "repo"}

	mock.ListWorkflows(context.Background(), repo)

	calls1 := mock.Calls()
	calls2 := mock.Calls()

	// Modify calls1
	if len(calls1) > 0 {
		calls1[0].Method = "Modified"
	}

	// calls2 should be unchanged
	if len(calls2) > 0 && calls2[0].Method == "Modified" {
		t.Error("Calls() should return a copy, not a reference")
	}
}

func TestMethodCall_Fields(t *testing.T) {
	now := time.Now()
	mc := MethodCall{
		Method: "TestMethod",
		Args:   []interface{}{"arg1", 42},
		Time:   now,
	}

	if mc.Method != "TestMethod" {
		t.Errorf("MethodCall.Method = %s, want TestMethod", mc.Method)
	}
	if len(mc.Args) != 2 {
		t.Errorf("len(MethodCall.Args) = %d, want 2", len(mc.Args))
	}
	if !mc.Time.Equal(now) {
		t.Errorf("MethodCall.Time = %v, want %v", mc.Time, now)
	}
}

func TestMockClient_ConcurrentAccess(t *testing.T) {
	mock := &MockClient{
		Workflows: []Workflow{{ID: 1, Name: "CI"}},
	}
	repo := Repository{Owner: "owner", Name: "repo"}

	// Test concurrent access is safe
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			mock.ListWorkflows(context.Background(), repo)
			mock.Calls()
			mock.CallCount("ListWorkflows")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if mock.CallCount("ListWorkflows") != 10 {
		t.Errorf("CallCount(ListWorkflows) = %d, want 10", mock.CallCount("ListWorkflows"))
	}
}
