package integration

import (
	"time"

	"github.com/nnnkkk7/lazyactions/github"
)

// DefaultTestWorkflows returns standard test workflows.
func DefaultTestWorkflows() []github.Workflow {
	return []github.Workflow{
		{ID: 1, Name: "CI", Path: ".github/workflows/ci.yml", State: "active"},
		{ID: 2, Name: "Deploy", Path: ".github/workflows/deploy.yml", State: "active"},
		{ID: 3, Name: "Release", Path: ".github/workflows/release.yml", State: "disabled"},
	}
}

// DefaultTestRuns returns standard test runs with various statuses.
func DefaultTestRuns() []github.Run {
	now := time.Now()
	return []github.Run{
		{
			ID:         100,
			Name:       "CI",
			Status:     "completed",
			Conclusion: "success",
			Branch:     "main",
			Event:      "push",
			Actor:      "user1",
			CreatedAt:  now.Add(-1 * time.Hour),
			URL:        "https://github.com/test/repo/actions/runs/100",
		},
		{
			ID:         101,
			Name:       "CI",
			Status:     "in_progress",
			Conclusion: "",
			Branch:     "feature/test",
			Event:      "pull_request",
			Actor:      "user2",
			CreatedAt:  now.Add(-30 * time.Minute),
			URL:        "https://github.com/test/repo/actions/runs/101",
		},
		{
			ID:         102,
			Name:       "CI",
			Status:     "completed",
			Conclusion: "failure",
			Branch:     "fix/bug",
			Event:      "push",
			Actor:      "user1",
			CreatedAt:  now.Add(-2 * time.Hour),
			URL:        "https://github.com/test/repo/actions/runs/102",
		},
		{
			ID:         103,
			Name:       "CI",
			Status:     "queued",
			Conclusion: "",
			Branch:     "main",
			Event:      "push",
			Actor:      "user3",
			CreatedAt:  now.Add(-5 * time.Minute),
			URL:        "https://github.com/test/repo/actions/runs/103",
		},
	}
}

// DefaultTestJobs returns standard test jobs.
func DefaultTestJobs() []github.Job {
	return []github.Job{
		{
			ID:         1001,
			Name:       "build",
			Status:     "completed",
			Conclusion: "success",
			Steps: []github.Step{
				{Name: "Checkout", Status: "completed", Conclusion: "success", Number: 1},
				{Name: "Setup Go", Status: "completed", Conclusion: "success", Number: 2},
				{Name: "Build", Status: "completed", Conclusion: "success", Number: 3},
			},
		},
		{
			ID:         1002,
			Name:       "test",
			Status:     "in_progress",
			Conclusion: "",
			Steps: []github.Step{
				{Name: "Checkout", Status: "completed", Conclusion: "success", Number: 1},
				{Name: "Setup Go", Status: "completed", Conclusion: "success", Number: 2},
				{Name: "Run Tests", Status: "in_progress", Conclusion: "", Number: 3},
			},
		},
		{
			ID:         1003,
			Name:       "lint",
			Status:     "completed",
			Conclusion: "failure",
			Steps: []github.Step{
				{Name: "Checkout", Status: "completed", Conclusion: "success", Number: 1},
				{Name: "Lint", Status: "completed", Conclusion: "failure", Number: 2},
			},
		},
	}
}

// DefaultTestLogs returns standard test logs.
func DefaultTestLogs() string {
	return `2024-01-15T10:00:00.000Z ##[group]Run actions/checkout@v4
2024-01-15T10:00:01.000Z with:
2024-01-15T10:00:01.000Z   repository: test/repo
2024-01-15T10:00:02.000Z ##[endgroup]
2024-01-15T10:00:03.000Z Syncing repository: test/repo
2024-01-15T10:00:05.000Z ##[group]Run go test -v ./...
2024-01-15T10:00:06.000Z === RUN   TestNew
2024-01-15T10:00:06.100Z --- PASS: TestNew (0.00s)
2024-01-15T10:00:06.200Z === RUN   TestApp_Init
2024-01-15T10:00:06.300Z --- PASS: TestApp_Init (0.00s)
2024-01-15T10:00:06.400Z PASS
2024-01-15T10:00:06.500Z ok      github.com/test/repo    0.010s
2024-01-15T10:00:06.600Z ##[endgroup]`
}

// FullMockData returns a complete mock dataset for testing.
func FullMockData() ([]github.Workflow, []github.Run, []github.Job, string) {
	return DefaultTestWorkflows(), DefaultTestRuns(), DefaultTestJobs(), DefaultTestLogs()
}

// RunningRun returns a run that is currently running.
func RunningRun() github.Run {
	return github.Run{
		ID:         200,
		Name:       "CI",
		Status:     "in_progress",
		Conclusion: "",
		Branch:     "main",
		Event:      "push",
		Actor:      "user1",
		CreatedAt:  time.Now().Add(-5 * time.Minute),
		URL:        "https://github.com/test/repo/actions/runs/200",
	}
}

// FailedRun returns a run that failed.
func FailedRun() github.Run {
	return github.Run{
		ID:         201,
		Name:       "CI",
		Status:     "completed",
		Conclusion: "failure",
		Branch:     "main",
		Event:      "push",
		Actor:      "user1",
		CreatedAt:  time.Now().Add(-10 * time.Minute),
		URL:        "https://github.com/test/repo/actions/runs/201",
	}
}

// SuccessRun returns a successful run.
func SuccessRun() github.Run {
	return github.Run{
		ID:         202,
		Name:       "CI",
		Status:     "completed",
		Conclusion: "success",
		Branch:     "main",
		Event:      "push",
		Actor:      "user1",
		CreatedAt:  time.Now().Add(-15 * time.Minute),
		URL:        "https://github.com/test/repo/actions/runs/202",
	}
}
