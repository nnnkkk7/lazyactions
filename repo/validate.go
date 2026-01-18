package repo

import (
	"fmt"
	"regexp"
	"strings"
)

// Validation errors
var (
	ErrInvalidOwnerName = fmt.Errorf("invalid owner name")
	ErrInvalidRepoName  = fmt.Errorf("invalid repository name")
)

// GitHub username/owner name rules:
// - Can only contain alphanumeric characters and hyphens
// - Cannot start or end with a hyphen
// - Cannot have consecutive hyphens
// - Maximum 39 characters
//
// We use a simple regex for basic character validation and additional checks
// for the hyphen rules since Go's regexp doesn't support lookaheads.
var validOwnerChars = regexp.MustCompile(`^[a-zA-Z0-9-]{1,39}$`)

// GitHub repository name rules:
// - Can contain alphanumeric characters, hyphens, underscores, and dots
// - Maximum 100 characters
// - Minimum 1 character
var validRepoName = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,100}$`)

// ValidateOwner validates a GitHub username/organization name.
// Returns an error if the name doesn't match GitHub's naming rules.
func ValidateOwner(owner string) error {
	if owner == "" {
		return fmt.Errorf("%w: cannot be empty", ErrInvalidOwnerName)
	}

	if !validOwnerChars.MatchString(owner) {
		return fmt.Errorf("%w: %q", ErrInvalidOwnerName, owner)
	}

	// Check hyphen rules: cannot start/end with hyphen or have consecutive hyphens
	if owner[0] == '-' || owner[len(owner)-1] == '-' || strings.Contains(owner, "--") {
		return fmt.Errorf("%w: %q", ErrInvalidOwnerName, owner)
	}

	return nil
}

// ValidateRepoName validates a GitHub repository name.
// Returns an error if the name doesn't match GitHub's naming rules.
func ValidateRepoName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: cannot be empty", ErrInvalidRepoName)
	}
	if !validRepoName.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrInvalidRepoName, name)
	}
	return nil
}

// ValidateRepository validates both owner and repository name.
// Returns the first validation error encountered.
func ValidateRepository(owner, name string) error {
	if err := ValidateOwner(owner); err != nil {
		return err
	}
	return ValidateRepoName(name)
}

// validWorkflowPath matches valid GitHub Actions workflow file paths.
var validWorkflowPath = regexp.MustCompile(`^\.github/workflows/[a-zA-Z0-9._-]+\.(yml|yaml)$`)

// ValidateWorkflowPath validates a GitHub Actions workflow file path.
// Valid paths must be in the format: .github/workflows/<name>.yml or .github/workflows/<name>.yaml
func ValidateWorkflowPath(path string) error {
	if path == "" {
		return fmt.Errorf("invalid workflow path: cannot be empty")
	}
	if !validWorkflowPath.MatchString(path) {
		return fmt.Errorf("invalid workflow path: %q", path)
	}
	return nil
}
