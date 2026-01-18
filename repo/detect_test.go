package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/nnnkkk7/lazyactions/github"
)

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected *github.Repository
		wantErr  bool
		errMsg   string
	}{
		// SSH format tests
		{
			name:     "SSH format basic",
			url:      "git@github.com:owner/repo.git",
			expected: &github.Repository{Owner: "owner", Name: "repo"},
			wantErr:  false,
		},
		{
			name:     "SSH format without .git suffix",
			url:      "git@github.com:owner/repo",
			expected: &github.Repository{Owner: "owner", Name: "repo"},
			wantErr:  false,
		},
		{
			name:     "SSH format with hyphens in owner and repo",
			url:      "git@github.com:my-org/my-repo.git",
			expected: &github.Repository{Owner: "my-org", Name: "my-repo"},
			wantErr:  false,
		},
		{
			name:     "SSH format with numbers",
			url:      "git@github.com:user123/project456.git",
			expected: &github.Repository{Owner: "user123", Name: "project456"},
			wantErr:  false,
		},
		{
			name:     "SSH format with underscores and dots",
			url:      "git@github.com:org_name/repo.name.git",
			expected: &github.Repository{Owner: "org_name", Name: "repo.name"},
			wantErr:  false,
		},
		{
			name:    "SSH format invalid - no repo",
			url:     "git@github.com:owner.git",
			wantErr: true,
			errMsg:  "invalid GitHub SSH URL",
		},
		{
			name:    "SSH format invalid - too many parts",
			url:     "git@github.com:owner/repo/extra.git",
			wantErr: true,
			errMsg:  "invalid GitHub SSH URL",
		},
		// HTTPS format tests
		{
			name:     "HTTPS format basic",
			url:      "https://github.com/owner/repo.git",
			expected: &github.Repository{Owner: "owner", Name: "repo"},
			wantErr:  false,
		},
		{
			name:     "HTTPS format without .git suffix",
			url:      "https://github.com/owner/repo",
			expected: &github.Repository{Owner: "owner", Name: "repo"},
			wantErr:  false,
		},
		{
			name:     "HTTPS format with hyphens in owner and repo",
			url:      "https://github.com/my-org/my-repo.git",
			expected: &github.Repository{Owner: "my-org", Name: "my-repo"},
			wantErr:  false,
		},
		{
			name:     "HTTPS format with numbers",
			url:      "https://github.com/user123/project456.git",
			expected: &github.Repository{Owner: "user123", Name: "project456"},
			wantErr:  false,
		},
		{
			name:     "HTTP format (insecure)",
			url:      "http://github.com/owner/repo.git",
			expected: &github.Repository{Owner: "owner", Name: "repo"},
			wantErr:  false,
		},
		{
			name:    "HTTPS format invalid - no repo",
			url:     "https://github.com/owner",
			wantErr: true,
			errMsg:  "invalid GitHub HTTPS URL",
		},
		{
			name:    "HTTPS format invalid - too many parts",
			url:     "https://github.com/owner/repo/extra",
			wantErr: true,
			errMsg:  "invalid GitHub HTTPS URL",
		},
		// Non-GitHub URLs
		{
			name:    "Non-GitHub URL - GitLab",
			url:     "git@gitlab.com:owner/repo.git",
			wantErr: true,
			errMsg:  "not a GitHub repository",
		},
		{
			name:    "Non-GitHub URL - Bitbucket",
			url:     "https://bitbucket.org/owner/repo.git",
			wantErr: true,
			errMsg:  "not a GitHub repository",
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "not a GitHub repository",
		},
		{
			name:    "Random string",
			url:     "not-a-url",
			wantErr: true,
			errMsg:  "not a GitHub repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGitHubURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseGitHubURL(%q) expected error, got nil", tt.url)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("parseGitHubURL(%q) error = %q, want error containing %q", tt.url, err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseGitHubURL(%q) unexpected error: %v", tt.url, err)
				return
			}

			if result.Owner != tt.expected.Owner {
				t.Errorf("parseGitHubURL(%q) Owner = %q, want %q", tt.url, result.Owner, tt.expected.Owner)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("parseGitHubURL(%q) Name = %q, want %q", tt.url, result.Name, tt.expected.Name)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	t.Run("not a git repository", func(t *testing.T) {
		// Create a temporary directory without .git
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		_, err := Detect()
		if err == nil {
			t.Error("Detect() expected error for non-git directory, got nil")
			return
		}
		if !contains(err.Error(), "not a git repository") {
			t.Errorf("Detect() error = %q, want error containing 'not a git repository'", err.Error())
		}
	})

	t.Run("git repository without remote", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// Initialize git repo without remote
		cmd := exec.Command("git", "init")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to initialize git repo: %v", err)
		}

		_, err := Detect()
		if err == nil {
			t.Error("Detect() expected error for git repo without remote, got nil")
			return
		}
		if !contains(err.Error(), "failed to get remote URL") {
			t.Errorf("Detect() error = %q, want error containing 'failed to get remote URL'", err.Error())
		}
	})

	t.Run("git repository with SSH remote", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// Initialize git repo and add SSH remote
		cmds := [][]string{
			{"git", "init"},
			{"git", "remote", "add", "origin", "git@github.com:testowner/testrepo.git"},
		}
		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to run %v: %v", args, err)
			}
		}

		result, err := Detect()
		if err != nil {
			t.Fatalf("Detect() unexpected error: %v", err)
		}

		if result.Owner != "testowner" {
			t.Errorf("Detect() Owner = %q, want %q", result.Owner, "testowner")
		}
		if result.Name != "testrepo" {
			t.Errorf("Detect() Name = %q, want %q", result.Name, "testrepo")
		}
	})

	t.Run("git repository with HTTPS remote", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// Initialize git repo and add HTTPS remote
		cmds := [][]string{
			{"git", "init"},
			{"git", "remote", "add", "origin", "https://github.com/httpsowner/httpsrepo.git"},
		}
		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to run %v: %v", args, err)
			}
		}

		result, err := Detect()
		if err != nil {
			t.Fatalf("Detect() unexpected error: %v", err)
		}

		if result.Owner != "httpsowner" {
			t.Errorf("Detect() Owner = %q, want %q", result.Owner, "httpsowner")
		}
		if result.Name != "httpsrepo" {
			t.Errorf("Detect() Name = %q, want %q", result.Name, "httpsrepo")
		}
	})

	t.Run("git repository with non-GitHub remote", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// Initialize git repo and add non-GitHub remote
		cmds := [][]string{
			{"git", "init"},
			{"git", "remote", "add", "origin", "git@gitlab.com:owner/repo.git"},
		}
		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to run %v: %v", args, err)
			}
		}

		_, err := Detect()
		if err == nil {
			t.Error("Detect() expected error for non-GitHub remote, got nil")
			return
		}
		if !contains(err.Error(), "not a GitHub repository") {
			t.Errorf("Detect() error = %q, want error containing 'not a GitHub repository'", err.Error())
		}
	})

	t.Run("nested directory in git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo with remote at root
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}
		cmds := [][]string{
			{"git", "init"},
			{"git", "remote", "add", "origin", "git@github.com:nested/project.git"},
		}
		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to run %v: %v", args, err)
			}
		}

		// Create nested directory and change to it
		nestedDir := filepath.Join(tmpDir, "src", "pkg")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("Failed to create nested directory: %v", err)
		}
		if err := os.Chdir(nestedDir); err != nil {
			t.Fatalf("Failed to change to nested directory: %v", err)
		}

		result, err := Detect()
		if err != nil {
			t.Fatalf("Detect() unexpected error: %v", err)
		}

		if result.Owner != "nested" {
			t.Errorf("Detect() Owner = %q, want %q", result.Owner, "nested")
		}
		if result.Name != "project" {
			t.Errorf("Detect() Name = %q, want %q", result.Name, "project")
		}
	})
}

func TestDetectFromPath(t *testing.T) {
	t.Run("valid git repository path", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo with remote
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		cmds := [][]string{
			{"git", "init"},
			{"git", "remote", "add", "origin", "git@github.com:pathowner/pathrepo.git"},
		}
		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Run()
		}
		os.Chdir(origDir)

		result, err := DetectFromPath(tmpDir)
		if err != nil {
			t.Fatalf("DetectFromPath(%q) unexpected error: %v", tmpDir, err)
		}

		if result.Owner != "pathowner" {
			t.Errorf("DetectFromPath() Owner = %q, want %q", result.Owner, "pathowner")
		}
		if result.Name != "pathrepo" {
			t.Errorf("DetectFromPath() Name = %q, want %q", result.Name, "pathrepo")
		}
	})

	t.Run("non-existent path", func(t *testing.T) {
		_, err := DetectFromPath("/non/existent/path")
		if err == nil {
			t.Error("DetectFromPath() expected error for non-existent path, got nil")
		}
	})

	t.Run("path without git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := DetectFromPath(tmpDir)
		if err == nil {
			t.Error("DetectFromPath() expected error for non-git directory, got nil")
		}
		if !contains(err.Error(), "not a git repository") {
			t.Errorf("DetectFromPath() error = %q, want error containing 'not a git repository'", err.Error())
		}
	})

	t.Run("returns to original directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo with remote
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		cmds := [][]string{
			{"git", "init"},
			{"git", "remote", "add", "origin", "git@github.com:owner/repo.git"},
		}
		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Run()
		}
		os.Chdir(origDir)

		beforeDir, _ := os.Getwd()
		DetectFromPath(tmpDir)
		afterDir, _ := os.Getwd()

		if beforeDir != afterDir {
			t.Errorf("DetectFromPath() changed working directory from %q to %q", beforeDir, afterDir)
		}
	})
}

// contains checks if s contains substr (helper function)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
