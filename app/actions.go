package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// User action functions - triggered by keyboard shortcuts

// confirmCancelRun shows confirmation dialog for cancelling a run
func (a *App) confirmCancelRun() tea.Cmd {
	run, ok := a.runs.Selected()
	if !ok || !run.IsRunning() {
		return nil
	}
	a.showConfirm = true
	a.confirmMsg = "Cancel this run?"
	a.confirmFn = func() tea.Cmd {
		return cancelRun(a.client, a.repo, run.ID)
	}
	return nil
}

// rerunWorkflow triggers a workflow rerun
func (a *App) rerunWorkflow() tea.Cmd {
	run, ok := a.runs.Selected()
	if !ok {
		return nil
	}
	return rerunWorkflow(a.client, a.repo, run.ID)
}

// rerunFailedJobs reruns only failed jobs
func (a *App) rerunFailedJobs() tea.Cmd {
	run, ok := a.runs.Selected()
	if !ok || !run.IsFailed() {
		return nil
	}
	return rerunFailedJobs(a.client, a.repo, run.ID)
}

// triggerWorkflow triggers a workflow dispatch
func (a *App) triggerWorkflow() tea.Cmd {
	wf, ok := a.workflows.Selected()
	if !ok {
		return nil
	}
	// Get workflow file name from path (e.g., ".github/workflows/ci.yml" -> "ci.yml")
	workflowFile := wf.Path
	if idx := len(".github/workflows/"); len(wf.Path) > idx {
		workflowFile = wf.Path[idx:]
	}
	// Trigger on default branch (main)
	return triggerWorkflow(a.client, a.repo, workflowFile, "main", nil)
}

// yankURL copies the selected run URL to clipboard
func (a *App) yankURL() tea.Cmd {
	run, ok := a.runs.Selected()
	if !ok || run.URL == "" {
		return nil
	}

	if err := a.clipboard.WriteAll(run.URL); err != nil {
		// Clipboard not available (e.g., headless environment)
		// Show URL in flash message so user can copy manually
		return flashMessage("URL: "+run.URL, FlashDurationInfo)
	}
	return flashMessage("Copied: "+run.URL, FlashDurationSuccess)
}

// refreshAll refreshes all data
func (a *App) refreshAll() tea.Cmd {
	a.loading = true
	return a.fetchWorkflowsCmd()
}

// refreshCurrentWorkflow refreshes runs for the current workflow
func (a *App) refreshCurrentWorkflow() tea.Cmd {
	if wf, ok := a.workflows.Selected(); ok {
		return a.fetchRunsCmd(wf.ID)
	}
	return nil
}
