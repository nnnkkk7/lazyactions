// Package repo provides repository detection and validation functionality.
package repo

import (
	"errors"
	"fmt"
	neturl "net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/nnnkkk7/lazyactions/github"
)

// ErrNotGitRepository is returned when the current directory is not a git repository.
var ErrNotGitRepository = errors.New("not a git repository")

// ErrNotGitHubRepository is returned when the remote URL is not a GitHub repository.
var ErrNotGitHubRepository = errors.New("not a GitHub repository")

// Detect detects the GitHub repository from the current directory.
// It reads the git remote origin URL and parses it to extract owner and repo name.
// Works from any subdirectory within a git repository.
func Detect() (*github.Repository, error) {
	// Check if we're in a git repository by running git rev-parse
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return nil, ErrNotGitRepository
	}

	// Get the remote origin URL
	cmd = exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	// Parse the URL and return the repository
	return parseGitHubURL(strings.TrimSpace(string(out)))
}

// parseGitHubURL parses a GitHub URL and extracts the owner and repository name.
// Supported formats:
//   - SSH: git@github.com:owner/repo.git
//   - HTTPS: https://github.com/owner/repo.git
//   - HTTP: http://github.com/owner/repo.git
//
// The .git suffix is optional in all formats.
func parseGitHubURL(url string) (*github.Repository, error) {
	// SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid GitHub SSH URL: %s", url)
		}
		return &github.Repository{Owner: parts[0], Name: parts[1]}, nil
	}

	// HTTPS/HTTP format: https://github.com/owner/repo.git
	if strings.Contains(url, "github.com") {
		u, err := neturl.Parse(url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}
		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid GitHub HTTPS URL: %s", url)
		}
		return &github.Repository{Owner: parts[0], Name: parts[1]}, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrNotGitHubRepository, url)
}

// DetectFromPath detects the GitHub repository from a specific path.
// It changes to the specified directory, detects the repository, then returns.
func DetectFromPath(path string) (*github.Repository, error) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Change to the specified path
	if err := os.Chdir(path); err != nil {
		return nil, fmt.Errorf("failed to change to directory %s: %w", path, err)
	}

	// Ensure we return to the original directory
	defer func() { _ = os.Chdir(origDir) }()

	// Detect the repository
	return Detect()
}
