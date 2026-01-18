package github

import (
	"strings"
	"testing"
)

func TestSanitizeLogs_GitHubTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "ghp_ classic PAT",
			input:    "Using token ghp_abcdefghijklmnopqrstuvwxyz0123456789 for auth",
			contains: "[REDACTED]",
		},
		{
			name:     "github_pat_ fine-grained PAT",
			input:    "Token: github_pat_1234567890abcdefghijkl_abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuv",
			contains: "[REDACTED]",
		},
		{
			name:     "gho_ OAuth token",
			input:    "OAuth gho_abcdefghijklmnopqrstuvwxyz0123456789",
			contains: "[REDACTED]",
		},
		{
			name:     "ghs_ server token",
			input:    "Server token: ghs_abcdefghijklmnopqrstuvwxyz0123456789",
			contains: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeLogs(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("SanitizeLogs() = %q, want to contain %q", result, tt.contains)
			}
			// Ensure original token is not present
			if strings.Contains(result, "ghp_") || strings.Contains(result, "github_pat_") ||
				strings.Contains(result, "gho_") || strings.Contains(result, "ghs_") {
				t.Errorf("SanitizeLogs() still contains token prefix in result: %q", result)
			}
		})
	}
}

func TestSanitizeLogs_AWSKeys(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "AWS access key ID",
			input: "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
		},
		{
			name:  "AWS secret access key",
			input: "aws_secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:  "AKIA prefix",
			input: "Key: AKIAIOSFODNN7EXAMPLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeLogs(tt.input)
			if !strings.Contains(result, "[REDACTED]") {
				t.Errorf("SanitizeLogs() = %q, want to contain [REDACTED]", result)
			}
		})
	}
}

func TestSanitizeLogs_GenericSecrets(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "api_key",
			input: "api_key=super_secret_key_12345678",
		},
		{
			name:  "apikey",
			input: "apikey: my-secret-api-key-value",
		},
		{
			name:  "password",
			input: "password=MySecretPassword123!",
		},
		{
			name:  "secret",
			input: "secret: very_secret_value_here",
		},
		{
			name:  "token",
			input: "token=abcdefghijklmnop",
		},
		{
			name:  "credential",
			input: "credential: user:password123",
		},
		{
			name:  "auth with quotes",
			input: `auth="secret_auth_token_value"`,
		},
		{
			name:  "API-KEY with hyphen",
			input: "API-KEY=my_secret_api_key_here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeLogs(tt.input)
			if !strings.Contains(result, "[REDACTED]") {
				t.Errorf("SanitizeLogs() = %q, want to contain [REDACTED]", result)
			}
		})
	}
}

func TestSanitizeLogs_NoSecrets(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "normal log line",
			input: "2024-01-15 10:30:00 INFO Starting build process",
		},
		{
			name:  "go test output",
			input: "=== RUN   TestFoo\n--- PASS: TestFoo (0.01s)",
		},
		{
			name:  "npm install output",
			input: "added 150 packages in 10s",
		},
		{
			name:  "short password value",
			input: "password=short", // Too short to match (< 8 chars)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeLogs(tt.input)
			if result != tt.input {
				t.Errorf("SanitizeLogs() = %q, want %q (no change expected)", result, tt.input)
			}
		})
	}
}

func TestSanitizeLogs_MultipleSecrets(t *testing.T) {
	input := `Starting deployment...
GitHub token: ghp_abcdefghijklmnopqrstuvwxyz0123456789
AWS key: AKIAIOSFODNN7EXAMPLE
password=SuperSecretPassword123
Done!`

	result := SanitizeLogs(input)

	// Count redactions
	redactCount := strings.Count(result, "[REDACTED]")
	if redactCount < 3 {
		t.Errorf("SanitizeLogs() redacted %d secrets, want at least 3", redactCount)
	}

	// Ensure secrets are removed
	if strings.Contains(result, "ghp_") {
		t.Error("SanitizeLogs() still contains GitHub token")
	}
	if strings.Contains(result, "AKIA") {
		t.Error("SanitizeLogs() still contains AWS key")
	}
}

func TestSanitizeLogs_EmptyInput(t *testing.T) {
	result := SanitizeLogs("")
	if result != "" {
		t.Errorf("SanitizeLogs(\"\") = %q, want \"\"", result)
	}
}

func TestContainsPotentialSecrets(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "contains GitHub token",
			content: "token: ghp_abcdefghijklmnopqrstuvwxyz0123456789",
			want:    true,
		},
		{
			name:    "contains AWS key",
			content: "AKIAIOSFODNN7EXAMPLE",
			want:    true,
		},
		{
			name:    "contains password",
			content: "password=secretvalue123",
			want:    true,
		},
		{
			name:    "no secrets",
			content: "Normal log output without any secrets",
			want:    false,
		},
		{
			name:    "empty string",
			content: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsPotentialSecrets(tt.content)
			if got != tt.want {
				t.Errorf("ContainsPotentialSecrets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeLogs_PreservesStructure(t *testing.T) {
	input := `Step 1: Setting up environment
  export API_KEY=my_super_secret_key_value
Step 2: Running tests
  go test ./...
Step 3: Complete`

	result := SanitizeLogs(input)

	// Should preserve line structure
	lines := strings.Split(result, "\n")
	if len(lines) != 5 {
		t.Errorf("SanitizeLogs() changed line count: got %d, want 5", len(lines))
	}

	// Should contain redaction
	if !strings.Contains(result, "[REDACTED]") {
		t.Error("SanitizeLogs() should contain [REDACTED]")
	}

	// Should preserve non-secret content
	if !strings.Contains(result, "Step 1") || !strings.Contains(result, "Step 3") {
		t.Error("SanitizeLogs() removed non-secret content")
	}
}

func BenchmarkSanitizeLogs(b *testing.B) {
	input := strings.Repeat("Normal log line without secrets\n", 100) +
		"ghp_abcdefghijklmnopqrstuvwxyz0123456789\n" +
		strings.Repeat("More normal log lines\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeLogs(input)
	}
}

func BenchmarkContainsPotentialSecrets(b *testing.B) {
	input := strings.Repeat("Normal log line without secrets\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ContainsPotentialSecrets(input)
	}
}
