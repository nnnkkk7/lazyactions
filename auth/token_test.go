package auth

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// =============================================================================
// SecureToken Tests
// =============================================================================

func TestSecureToken_String_ReturnsRedacted(t *testing.T) {
	token, err := NewSecureToken("ghp_1234567890abcdef1234567890abcdef12345")
	if err != nil {
		t.Fatalf("NewSecureToken failed: %v", err)
	}

	got := token.String()
	want := "[REDACTED]"

	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestSecureToken_GoString_ReturnsRedacted(t *testing.T) {
	token, err := NewSecureToken("ghp_1234567890abcdef1234567890abcdef12345")
	if err != nil {
		t.Fatalf("NewSecureToken failed: %v", err)
	}

	got := token.GoString()
	want := "[REDACTED]"

	if got != want {
		t.Errorf("GoString() = %q, want %q", got, want)
	}
}

func TestSecureToken_Value_ReturnsActualToken(t *testing.T) {
	tokenValue := "ghp_1234567890abcdef1234567890abcdef12345"
	token, err := NewSecureToken(tokenValue)
	if err != nil {
		t.Fatalf("NewSecureToken failed: %v", err)
	}

	got := token.Value()

	if got != tokenValue {
		t.Errorf("Value() = %q, want %q", got, tokenValue)
	}
}

func TestSecureToken_FmtPrintf_HidesToken(t *testing.T) {
	token, err := NewSecureToken("ghp_1234567890abcdef1234567890abcdef12345")
	if err != nil {
		t.Fatalf("NewSecureToken failed: %v", err)
	}

	tests := []struct {
		name   string
		format string
	}{
		{"percent_s", "%s"},
		{"percent_v", "%v"},
		{"percent_plus_v", "%+v"},
		{"percent_hash_v", "%#v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmt.Sprintf(tt.format, token)
			if strings.Contains(got, "ghp_") {
				t.Errorf("fmt.Sprintf(%q, token) = %q, should not contain actual token", tt.format, got)
			}
			if !strings.Contains(got, "[REDACTED]") {
				t.Errorf("fmt.Sprintf(%q, token) = %q, should contain [REDACTED]", tt.format, got)
			}
		})
	}
}

// =============================================================================
// NewSecureToken Validation Tests
// =============================================================================

func TestNewSecureToken_ValidTokenFormats(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"classic_PAT_ghp", "ghp_1234567890abcdef1234567890abcdef12345"},
		{"fine_grained_PAT_github_pat", "github_pat_11ABCDEFG_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890a"},
		{"OAuth_token_gho", "gho_1234567890abcdef1234567890abcdef12345"},
		{"server_to_server_ghs", "ghs_1234567890abcdef1234567890abcdef12345"},
		{"legacy_40_char_hex", "1234567890abcdef1234567890abcdef12345678"},
		{"legacy_40_char_hex_upper", "1234567890ABCDEF1234567890ABCDEF12345678"},
		{"legacy_40_char_hex_mixed", "1234567890AbCdEf1234567890AbCdEf12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewSecureToken(tt.token)
			if err != nil {
				t.Errorf("NewSecureToken(%q) returned error: %v", tt.token, err)
				return
			}
			if token.Value() != tt.token {
				t.Errorf("token.Value() = %q, want %q", token.Value(), tt.token)
			}
		})
	}
}

func TestNewSecureToken_InvalidTokenFormats(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		wantErrMsg string
	}{
		{"empty_string", "", "token is empty"},
		{"too_short", "abc123", "invalid token format"},
		{"wrong_prefix", "xyz_1234567890abcdef1234567890abcdef12345", "invalid token format"},
		{"39_chars_hex", "123456789012345678901234567890123456789", "invalid token format"},
		{"41_chars_hex", "12345678901234567890123456789012345678901", "invalid token format"},
		{"non_hex_40_chars", "gggggggggggggggggggggggggggggggggggggggg", "invalid token format"},
		{"whitespace_only", "   ", "invalid token format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSecureToken(tt.token)
			if err == nil {
				t.Errorf("NewSecureToken(%q) should return error", tt.token)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("NewSecureToken(%q) error = %q, want to contain %q", tt.token, err.Error(), tt.wantErrMsg)
			}
		})
	}
}

// =============================================================================
// GetToken Priority Tests
// =============================================================================

func TestGetToken_PrioritizesGhCLI(t *testing.T) {
	// Skip if gh CLI is not available
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not installed, skipping gh CLI priority test")
	}

	// Set up environment variable as fallback
	originalEnv := os.Getenv("GITHUB_TOKEN")
	os.Setenv("GITHUB_TOKEN", "ghp_envtoken1234567890abcdef1234567890ab")
	defer os.Setenv("GITHUB_TOKEN", originalEnv)

	// Try to get token from gh CLI
	ghToken, err := getFromGhCLI()
	if err != nil {
		t.Skip("gh CLI not authenticated, skipping priority test")
	}

	// Get token using main function
	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() returned error: %v", err)
	}

	// Should get the gh CLI token, not the env var
	if token.Value() != ghToken {
		t.Errorf("GetToken() = %q (from env), want %q (from gh CLI)", token.Value(), ghToken)
	}
}

func TestGetToken_FallsBackToEnvVar(t *testing.T) {
	// Mock gh CLI by setting a function that returns error
	originalGetFromGhCLI := ghCLIFunc
	ghCLIFunc = func() (string, error) {
		return "", fmt.Errorf("gh CLI not available")
	}
	defer func() { ghCLIFunc = originalGetFromGhCLI }()

	// Set up environment variable
	originalEnv := os.Getenv("GITHUB_TOKEN")
	envToken := "ghp_envtoken1234567890abcdef1234567890ab"
	os.Setenv("GITHUB_TOKEN", envToken)
	defer os.Setenv("GITHUB_TOKEN", originalEnv)

	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() returned error: %v", err)
	}

	if token.Value() != envToken {
		t.Errorf("GetToken() = %q, want %q", token.Value(), envToken)
	}
}

func TestGetToken_ReturnsErrorWhenNoTokenAvailable(t *testing.T) {
	// Mock gh CLI to return error
	originalGetFromGhCLI := ghCLIFunc
	ghCLIFunc = func() (string, error) {
		return "", fmt.Errorf("gh CLI not available")
	}
	defer func() { ghCLIFunc = originalGetFromGhCLI }()

	// Clear environment variable
	originalEnv := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalEnv)

	_, err := GetToken()
	if err == nil {
		t.Error("GetToken() should return error when no token available")
	}

	expectedMsg := "no authentication token found"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("GetToken() error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestGetToken_GhCLI_EmptyTokenFallsBackToEnv(t *testing.T) {
	// Mock gh CLI to return empty string
	originalGetFromGhCLI := ghCLIFunc
	ghCLIFunc = func() (string, error) {
		return "", nil
	}
	defer func() { ghCLIFunc = originalGetFromGhCLI }()

	// Set up environment variable
	originalEnv := os.Getenv("GITHUB_TOKEN")
	envToken := "ghp_envtoken1234567890abcdef1234567890ab"
	os.Setenv("GITHUB_TOKEN", envToken)
	defer os.Setenv("GITHUB_TOKEN", originalEnv)

	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() returned error: %v", err)
	}

	if token.Value() != envToken {
		t.Errorf("GetToken() = %q, want %q (fallback to env)", token.Value(), envToken)
	}
}

// =============================================================================
// getFromGhCLI Tests
// =============================================================================

func TestGetFromGhCLI_ReturnsTokenWhenAuthenticated(t *testing.T) {
	// Skip if gh CLI is not available
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not installed")
	}

	token, err := getFromGhCLI()
	if err != nil {
		t.Skip("gh CLI not authenticated")
	}

	if token == "" {
		t.Error("getFromGhCLI() returned empty token when authenticated")
	}
}

func TestGetFromGhCLI_TrimsWhitespace(t *testing.T) {
	// This test verifies the implementation trims whitespace
	// We can't directly test this without mocking, but we verify
	// through the design that strings.TrimSpace is used
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not installed")
	}

	token, err := getFromGhCLI()
	if err != nil {
		t.Skip("gh CLI not authenticated")
	}

	// Token should not have leading/trailing whitespace
	if token != strings.TrimSpace(token) {
		t.Errorf("getFromGhCLI() returned token with whitespace: %q", token)
	}
}

// =============================================================================
// isValidGitHubTokenFormat Tests
// =============================================================================

func TestIsValidGitHubTokenFormat(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		// Valid formats
		{"ghp_prefix", "ghp_abcdefghijklmnopqrstuvwxyz1234567890", true},
		{"github_pat_prefix", "github_pat_11ABC_xyz", true},
		{"gho_prefix", "gho_abcdefghijklmnopqrstuvwxyz1234567890", true},
		{"ghs_prefix", "ghs_abcdefghijklmnopqrstuvwxyz1234567890", true},
		{"40_char_hex_lower", "abcdef1234567890abcdef1234567890abcdef12", true},
		{"40_char_hex_upper", "ABCDEF1234567890ABCDEF1234567890ABCDEF12", true},
		{"40_char_hex_mixed", "AbCdEf1234567890AbCdEf1234567890AbCdEf12", true},

		// Invalid formats
		{"empty", "", false},
		{"short", "abc", false},
		{"wrong_prefix", "xyz_token", false},
		{"39_chars", "abcdef1234567890abcdef1234567890abcdef1", false},
		{"41_chars", "abcdef1234567890abcdef1234567890abcdef123", false},
		{"40_chars_non_hex", "ghijklmnopqrstuvwxyzghijklmnopqrstuvwxyz", false},
		{"has_spaces", "ghp_ token", true}, // prefix match, rest doesn't matter
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidGitHubTokenFormat(tt.token)
			if got != tt.want {
				t.Errorf("isValidGitHubTokenFormat(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

// =============================================================================
// Edge Cases and Integration Tests
// =============================================================================

func TestSecureToken_ZeroValue(t *testing.T) {
	var token SecureToken

	// Zero value should still return [REDACTED]
	if token.String() != "[REDACTED]" {
		t.Errorf("zero value String() = %q, want [REDACTED]", token.String())
	}

	if token.GoString() != "[REDACTED]" {
		t.Errorf("zero value GoString() = %q, want [REDACTED]", token.GoString())
	}

	// Zero value Value() should return empty string
	if token.Value() != "" {
		t.Errorf("zero value Value() = %q, want empty string", token.Value())
	}
}

func TestGetToken_ValidatesTokenFormat(t *testing.T) {
	// Mock gh CLI to return invalid token
	originalGetFromGhCLI := ghCLIFunc
	ghCLIFunc = func() (string, error) {
		return "invalid_token", nil
	}
	defer func() { ghCLIFunc = originalGetFromGhCLI }()

	// Clear environment variable
	originalEnv := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalEnv)

	_, err := GetToken()
	if err == nil {
		t.Error("GetToken() should return error for invalid token format")
	}
}

func TestGetToken_ValidatesEnvVarTokenFormat(t *testing.T) {
	// Mock gh CLI to return error (so we fall back to env)
	originalGetFromGhCLI := ghCLIFunc
	ghCLIFunc = func() (string, error) {
		return "", fmt.Errorf("gh CLI not available")
	}
	defer func() { ghCLIFunc = originalGetFromGhCLI }()

	// Set invalid token in environment variable
	originalEnv := os.Getenv("GITHUB_TOKEN")
	os.Setenv("GITHUB_TOKEN", "invalid_token")
	defer os.Setenv("GITHUB_TOKEN", originalEnv)

	_, err := GetToken()
	if err == nil {
		t.Error("GetToken() should return error for invalid env token format")
	}
}
