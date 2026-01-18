package github

import (
	"testing"
	"time"
)

func TestRepository_FullName(t *testing.T) {
	tests := []struct {
		name string
		repo Repository
		want string
	}{
		{
			name: "standard repository",
			repo: Repository{Owner: "owner", Name: "repo"},
			want: "owner/repo",
		},
		{
			name: "organization repository",
			repo: Repository{Owner: "my-org", Name: "my-project"},
			want: "my-org/my-project",
		},
		{
			name: "empty values",
			repo: Repository{Owner: "", Name: ""},
			want: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.repo.FullName()
			if got != tt.want {
				t.Errorf("Repository.FullName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun_IsRunning(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{name: "queued", status: "queued", want: true},
		{name: "in_progress", status: "in_progress", want: true},
		{name: "completed", status: "completed", want: false},
		{name: "cancelled", status: "cancelled", want: false},
		{name: "empty", status: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Run{Status: tt.status}
			got := r.IsRunning()
			if got != tt.want {
				t.Errorf("Run.IsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun_IsFailed(t *testing.T) {
	tests := []struct {
		name       string
		conclusion string
		want       bool
	}{
		{name: "failure", conclusion: "failure", want: true},
		{name: "success", conclusion: "success", want: false},
		{name: "cancelled", conclusion: "cancelled", want: false},
		{name: "skipped", conclusion: "skipped", want: false},
		{name: "empty", conclusion: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Run{Conclusion: tt.conclusion}
			got := r.IsFailed()
			if got != tt.want {
				t.Errorf("Run.IsFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStep_IsFailed(t *testing.T) {
	tests := []struct {
		name       string
		conclusion string
		want       bool
	}{
		{name: "failure", conclusion: "failure", want: true},
		{name: "success", conclusion: "success", want: false},
		{name: "skipped", conclusion: "skipped", want: false},
		{name: "empty", conclusion: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Step{Conclusion: tt.conclusion}
			got := s.IsFailed()
			if got != tt.want {
				t.Errorf("Step.IsFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkflow_Fields(t *testing.T) {
	w := Workflow{
		ID:    12345,
		Name:  "CI",
		Path:  ".github/workflows/ci.yml",
		State: "active",
	}

	if w.ID != 12345 {
		t.Errorf("Workflow.ID = %v, want 12345", w.ID)
	}
	if w.Name != "CI" {
		t.Errorf("Workflow.Name = %v, want CI", w.Name)
	}
	if w.Path != ".github/workflows/ci.yml" {
		t.Errorf("Workflow.Path = %v, want .github/workflows/ci.yml", w.Path)
	}
	if w.State != "active" {
		t.Errorf("Workflow.State = %v, want active", w.State)
	}
}

func TestRun_Fields(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	r := Run{
		ID:         100,
		Name:       "CI",
		Status:     "completed",
		Conclusion: "success",
		Branch:     "main",
		Event:      "push",
		CreatedAt:  createdAt,
		Actor:      "user1",
		URL:        "https://github.com/owner/repo/actions/runs/100",
	}

	if r.ID != 100 {
		t.Errorf("Run.ID = %v, want 100", r.ID)
	}
	if r.Name != "CI" {
		t.Errorf("Run.Name = %v, want CI", r.Name)
	}
	if r.Status != "completed" {
		t.Errorf("Run.Status = %v, want completed", r.Status)
	}
	if r.Conclusion != "success" {
		t.Errorf("Run.Conclusion = %v, want success", r.Conclusion)
	}
	if r.Branch != "main" {
		t.Errorf("Run.Branch = %v, want main", r.Branch)
	}
	if r.Event != "push" {
		t.Errorf("Run.Event = %v, want push", r.Event)
	}
	if !r.CreatedAt.Equal(createdAt) {
		t.Errorf("Run.CreatedAt = %v, want %v", r.CreatedAt, createdAt)
	}
	if r.Actor != "user1" {
		t.Errorf("Run.Actor = %v, want user1", r.Actor)
	}
	if r.URL != "https://github.com/owner/repo/actions/runs/100" {
		t.Errorf("Run.URL = %v, want https://github.com/owner/repo/actions/runs/100", r.URL)
	}
}

func TestJob_Fields(t *testing.T) {
	steps := []Step{
		{Name: "Checkout", Status: "completed", Conclusion: "success", Number: 1},
		{Name: "Build", Status: "completed", Conclusion: "failure", Number: 2},
	}
	j := Job{
		ID:         200,
		Name:       "build",
		Status:     "completed",
		Conclusion: "failure",
		Steps:      steps,
	}

	if j.ID != 200 {
		t.Errorf("Job.ID = %v, want 200", j.ID)
	}
	if j.Name != "build" {
		t.Errorf("Job.Name = %v, want build", j.Name)
	}
	if j.Status != "completed" {
		t.Errorf("Job.Status = %v, want completed", j.Status)
	}
	if j.Conclusion != "failure" {
		t.Errorf("Job.Conclusion = %v, want failure", j.Conclusion)
	}
	if len(j.Steps) != 2 {
		t.Errorf("len(Job.Steps) = %v, want 2", len(j.Steps))
	}
	if j.Steps[0].Name != "Checkout" {
		t.Errorf("Job.Steps[0].Name = %v, want Checkout", j.Steps[0].Name)
	}
	if !j.Steps[1].IsFailed() {
		t.Errorf("Job.Steps[1].IsFailed() = false, want true")
	}
}

func TestStep_Fields(t *testing.T) {
	s := Step{
		Name:       "Run tests",
		Status:     "completed",
		Conclusion: "success",
		Number:     3,
	}

	if s.Name != "Run tests" {
		t.Errorf("Step.Name = %v, want Run tests", s.Name)
	}
	if s.Status != "completed" {
		t.Errorf("Step.Status = %v, want completed", s.Status)
	}
	if s.Conclusion != "success" {
		t.Errorf("Step.Conclusion = %v, want success", s.Conclusion)
	}
	if s.Number != 3 {
		t.Errorf("Step.Number = %v, want 3", s.Number)
	}
}

func TestListRunsOpts_Fields(t *testing.T) {
	opts := ListRunsOpts{
		WorkflowID: 123,
		Branch:     "main",
		Event:      "push",
		Status:     "completed",
		PerPage:    50,
	}

	if opts.WorkflowID != 123 {
		t.Errorf("ListRunsOpts.WorkflowID = %v, want 123", opts.WorkflowID)
	}
	if opts.Branch != "main" {
		t.Errorf("ListRunsOpts.Branch = %v, want main", opts.Branch)
	}
	if opts.Event != "push" {
		t.Errorf("ListRunsOpts.Event = %v, want push", opts.Event)
	}
	if opts.Status != "completed" {
		t.Errorf("ListRunsOpts.Status = %v, want completed", opts.Status)
	}
	if opts.PerPage != 50 {
		t.Errorf("ListRunsOpts.PerPage = %v, want 50", opts.PerPage)
	}
}
