// Package auth provides GitHub authentication functionality.
// It supports token retrieval from gh CLI and environment variables.
package auth

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ghCLIFunc is the function used to get token from gh CLI.
// It can be replaced in tests for mocking.
var ghCLIFunc = getFromGhCLI

// legacyTokenPattern matches the 40-character hexadecimal legacy token format.
var legacyTokenPattern = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

// SecureToken wraps a GitHub token and prevents accidental exposure in logs.
// It implements fmt.Stringer and fmt.GoStringer to return "[REDACTED]"
// instead of the actual token value.
type SecureToken struct {
	value string
}

// String returns "[REDACTED]" to prevent token leakage in logs.
// This implements the fmt.Stringer interface.
func (t SecureToken) String() string {
	return "[REDACTED]"
}

// GoString returns "[REDACTED]" to prevent token leakage in %#v format.
// This implements the fmt.GoStringer interface.
func (t SecureToken) GoString() string {
	return "[REDACTED]"
}

// Value returns the actual token value.
// Use this method only when the token needs to be passed to the GitHub API.
func (t SecureToken) Value() string {
	return t.value
}

// NewSecureToken creates a new SecureToken after validating the token format.
// It returns an error if the token is empty or has an invalid format.
func NewSecureToken(token string) (SecureToken, error) {
	if token == "" {
		return SecureToken{}, errors.New("token is empty")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return SecureToken{}, errors.New("invalid token format")
	}

	if !isValidGitHubTokenFormat(token) {
		return SecureToken{}, errors.New("invalid token format")
	}

	return SecureToken{value: token}, nil
}

// GetToken retrieves a GitHub token using the following priority:
//  1. gh CLI (gh auth token) - recommended, no configuration needed
//  2. GITHUB_TOKEN environment variable
//
// Returns an error if no valid token is found.
func GetToken() (SecureToken, error) {
	// 1. Try gh CLI first (highest priority)
	if token, err := ghCLIFunc(); err == nil && token != "" {
		return NewSecureToken(token)
	}

	// 2. Fall back to environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return NewSecureToken(token)
	}

	return SecureToken{}, errors.New("no authentication token found: run 'gh auth login' or set GITHUB_TOKEN")
}

// getFromGhCLI retrieves the GitHub token from the gh CLI.
// It runs "gh auth token" and returns the output.
func getFromGhCLI() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// isValidGitHubTokenFormat checks if the token has a valid GitHub token format.
// Valid formats include:
//   - ghp_* (classic personal access token)
//   - github_pat_* (fine-grained personal access token)
//   - gho_* (OAuth token)
//   - ghs_* (server-to-server token)
//   - 40 character hexadecimal string (legacy format)
func isValidGitHubTokenFormat(token string) bool {
	if token == "" {
		return false
	}

	// Check for known prefixes
	validPrefixes := []string{"ghp_", "github_pat_", "gho_", "ghs_"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(token, prefix) {
			return true
		}
	}

	// Check for legacy 40-character hex format
	if len(token) == 40 {
		return legacyTokenPattern.MatchString(token)
	}

	return false
}
