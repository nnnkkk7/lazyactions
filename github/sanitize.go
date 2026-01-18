// Package github provides sanitization utilities for log content.
package github

import (
	"regexp"
)

// secretPatterns contains regex patterns for detecting secrets in logs.
var secretPatterns = []*regexp.Regexp{
	// GitHub tokens
	regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
	regexp.MustCompile(`github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59}`),
	regexp.MustCompile(`gho_[a-zA-Z0-9]{36}`),
	regexp.MustCompile(`ghs_[a-zA-Z0-9]{36}`),
	// AWS
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	regexp.MustCompile(`(?i)(aws_secret_access_key|aws_access_key_id)\s*[=:]\s*['"]?[A-Za-z0-9/+=]{20,}['"]?`),
	// Generic secrets
	regexp.MustCompile(`(?i)(api[_-]?key|apikey|secret|password|token|credential|auth)[=:]\s*['"]?[^\s'"]{8,}['"]?`),
}

// SanitizeLogs removes potential secrets from log content.
// It replaces matched patterns with "[REDACTED]".
func SanitizeLogs(logs string) string {
	result := logs
	for _, pattern := range secretPatterns {
		result = pattern.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}

// ContainsPotentialSecrets checks if the content contains potential secrets.
// Returns true if any secret pattern matches.
func ContainsPotentialSecrets(content string) bool {
	for _, pattern := range secretPatterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
}
