package github

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v68/github"
)

// realClient implements the Client interface using go-github
type realClient struct {
	client    *github.Client
	owner     string
	repoName  string
	rateLimit int
}

// NewClient creates a new GitHub API client
func NewClient(token, owner, repoName string) Client {
	httpClient := &http.Client{}
	if token != "" {
		httpClient = &http.Client{
			Transport: &tokenTransport{token: token},
		}
	}

	return &realClient{
		client:    github.NewClient(httpClient),
		owner:     owner,
		repoName:  repoName,
		rateLimit: 5000, // Default rate limit
	}
}

// tokenTransport adds authorization header to requests
type tokenTransport struct {
	token string
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(req)
}

// updateRateLimit updates the rate limit from the response.
func (c *realClient) updateRateLimit(resp *github.Response) {
	if resp != nil {
		c.rateLimit = resp.Rate.Remaining
	}
}

// ListWorkflows lists all workflows in the repository.
func (c *realClient) ListWorkflows(ctx context.Context, repo Repository) ([]Workflow, error) {
	opts := &github.ListOptions{PerPage: 100}
	workflows, resp, err := c.client.Actions.ListWorkflows(ctx, repo.Owner, repo.Name, opts)
	c.updateRateLimit(resp)
	if err != nil {
		return nil, WrapAPIError(err)
	}

	result := make([]Workflow, 0, len(workflows.Workflows))
	for _, w := range workflows.Workflows {
		result = append(result, Workflow{
			ID:    w.GetID(),
			Name:  w.GetName(),
			Path:  w.GetPath(),
			State: w.GetState(),
		})
	}
	return result, nil
}

// ListRuns lists workflow runs.
func (c *realClient) ListRuns(ctx context.Context, repo Repository, opts *ListRunsOpts) ([]Run, error) {
	ghOpts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 30},
	}
	if opts != nil {
		if opts.PerPage > 0 {
			ghOpts.ListOptions.PerPage = opts.PerPage
		}
		if opts.Branch != "" {
			ghOpts.Branch = opts.Branch
		}
		if opts.Status != "" {
			ghOpts.Status = opts.Status
		}
		if opts.Event != "" {
			ghOpts.Event = opts.Event
		}
		if opts.WorkflowID > 0 {
			runs, resp, err := c.client.Actions.ListWorkflowRunsByID(ctx, repo.Owner, repo.Name, opts.WorkflowID, ghOpts)
			c.updateRateLimit(resp)
			if err != nil {
				return nil, WrapAPIError(err)
			}
			return convertRuns(runs.WorkflowRuns), nil
		}
	}

	runs, resp, err := c.client.Actions.ListRepositoryWorkflowRuns(ctx, repo.Owner, repo.Name, ghOpts)
	c.updateRateLimit(resp)
	if err != nil {
		return nil, WrapAPIError(err)
	}
	return convertRuns(runs.WorkflowRuns), nil
}

// CancelRun cancels a workflow run.
func (c *realClient) CancelRun(ctx context.Context, repo Repository, runID int64) error {
	resp, err := c.client.Actions.CancelWorkflowRunByID(ctx, repo.Owner, repo.Name, runID)
	c.updateRateLimit(resp)
	if err != nil {
		return WrapAPIError(err)
	}
	return nil
}

// RerunWorkflow reruns a workflow.
func (c *realClient) RerunWorkflow(ctx context.Context, repo Repository, runID int64) error {
	resp, err := c.client.Actions.RerunWorkflowByID(ctx, repo.Owner, repo.Name, runID)
	c.updateRateLimit(resp)
	if err != nil {
		return WrapAPIError(err)
	}
	return nil
}

// RerunFailedJobs reruns only failed jobs in a workflow.
func (c *realClient) RerunFailedJobs(ctx context.Context, repo Repository, runID int64) error {
	resp, err := c.client.Actions.RerunFailedJobsByID(ctx, repo.Owner, repo.Name, runID)
	c.updateRateLimit(resp)
	if err != nil {
		return WrapAPIError(err)
	}
	return nil
}

// TriggerWorkflow triggers a workflow_dispatch event.
func (c *realClient) TriggerWorkflow(ctx context.Context, repo Repository, workflowFile, ref string, inputs map[string]interface{}) error {
	event := github.CreateWorkflowDispatchEventRequest{
		Ref:    ref,
		Inputs: inputs,
	}
	resp, err := c.client.Actions.CreateWorkflowDispatchEventByFileName(ctx, repo.Owner, repo.Name, workflowFile, event)
	c.updateRateLimit(resp)
	if err != nil {
		return WrapAPIError(err)
	}
	return nil
}

// ListJobs lists jobs for a workflow run.
func (c *realClient) ListJobs(ctx context.Context, repo Repository, runID int64) ([]Job, error) {
	opts := &github.ListWorkflowJobsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	jobs, resp, err := c.client.Actions.ListWorkflowJobs(ctx, repo.Owner, repo.Name, runID, opts)
	c.updateRateLimit(resp)
	if err != nil {
		return nil, WrapAPIError(err)
	}

	result := make([]Job, 0, len(jobs.Jobs))
	for _, j := range jobs.Jobs {
		steps := make([]Step, 0, len(j.Steps))
		for _, s := range j.Steps {
			steps = append(steps, Step{
				Name:       s.GetName(),
				Status:     s.GetStatus(),
				Conclusion: s.GetConclusion(),
				Number:     int(s.GetNumber()),
			})
		}
		result = append(result, Job{
			ID:         j.GetID(),
			Name:       j.GetName(),
			Status:     j.GetStatus(),
			Conclusion: j.GetConclusion(),
			Steps:      steps,
		})
	}
	return result, nil
}

// GetJobLogs gets logs for a job.
func (c *realClient) GetJobLogs(ctx context.Context, repo Repository, jobID int64) (string, error) {
	url, resp, err := c.client.Actions.GetWorkflowJobLogs(ctx, repo.Owner, repo.Name, jobID, 2)
	c.updateRateLimit(resp)
	if err != nil {
		return "", WrapAPIError(err)
	}

	logResp, err := http.Get(url.String())
	if err != nil {
		return "", fmt.Errorf("failed to download logs: %w", err)
	}
	defer func() { _ = logResp.Body.Close() }()

	body, err := io.ReadAll(logResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return string(body), nil
}

// RateLimitRemaining returns the remaining rate limit.
func (c *realClient) RateLimitRemaining() int {
	return c.rateLimit
}

// convertRuns converts GitHub API runs to our Run type.
func convertRuns(ghRuns []*github.WorkflowRun) []Run {
	result := make([]Run, 0, len(ghRuns))
	for _, r := range ghRuns {
		result = append(result, Run{
			ID:         r.GetID(),
			Name:       r.GetName(),
			Status:     r.GetStatus(),
			Conclusion: r.GetConclusion(),
			Branch:     r.GetHeadBranch(),
			Event:      r.GetEvent(),
			Actor:      r.GetActor().GetLogin(),
			URL:        r.GetHTMLURL(),
			CreatedAt:  r.GetCreatedAt().Time,
		})
	}
	return result
}
