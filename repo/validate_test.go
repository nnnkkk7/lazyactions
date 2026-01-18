package repo

import "testing"

func TestValidateOwner(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		wantErr bool
		errMsg  string
	}{
		// Valid owners
		{
			name:    "simple lowercase",
			owner:   "owner",
			wantErr: false,
		},
		{
			name:    "simple uppercase",
			owner:   "Owner",
			wantErr: false,
		},
		{
			name:    "mixed case",
			owner:   "MyOrg",
			wantErr: false,
		},
		{
			name:    "with numbers",
			owner:   "user123",
			wantErr: false,
		},
		{
			name:    "starting with number",
			owner:   "123user",
			wantErr: false,
		},
		{
			name:    "with hyphen in middle",
			owner:   "my-org",
			wantErr: false,
		},
		{
			name:    "multiple hyphens",
			owner:   "my-cool-org",
			wantErr: false,
		},
		{
			name:    "single character",
			owner:   "a",
			wantErr: false,
		},
		{
			name:    "max length 39 chars",
			owner:   "a23456789012345678901234567890123456789",
			wantErr: false,
		},
		// Invalid owners
		{
			name:    "empty string",
			owner:   "",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name:    "starting with hyphen",
			owner:   "-owner",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name:    "ending with hyphen",
			owner:   "owner-",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name:    "consecutive hyphens",
			owner:   "my--org",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name:    "with underscore",
			owner:   "my_org",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name:    "with dot",
			owner:   "my.org",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name:    "with space",
			owner:   "my org",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name:    "with special characters",
			owner:   "my@org",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name:    "too long (40 chars)",
			owner:   "a234567890123456789012345678901234567890",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name:    "with unicode",
			owner:   "my-org-\u65e5\u672c",
			wantErr: true,
			errMsg:  "invalid owner name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOwner(tt.owner)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateOwner(%q) expected error, got nil", tt.owner)
					return
				}
				if tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("ValidateOwner(%q) error = %q, want error containing %q", tt.owner, err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateOwner(%q) unexpected error: %v", tt.owner, err)
			}
		})
	}
}

func TestValidateRepoName(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		wantErr  bool
		errMsg   string
	}{
		// Valid repo names
		{
			name:     "simple lowercase",
			repoName: "myrepo",
			wantErr:  false,
		},
		{
			name:     "simple uppercase",
			repoName: "MyRepo",
			wantErr:  false,
		},
		{
			name:     "mixed case",
			repoName: "MyAwesomeRepo",
			wantErr:  false,
		},
		{
			name:     "with numbers",
			repoName: "repo123",
			wantErr:  false,
		},
		{
			name:     "starting with number",
			repoName: "123repo",
			wantErr:  false,
		},
		{
			name:     "with hyphen",
			repoName: "my-repo",
			wantErr:  false,
		},
		{
			name:     "with underscore",
			repoName: "my_repo",
			wantErr:  false,
		},
		{
			name:     "with dot",
			repoName: "my.repo",
			wantErr:  false,
		},
		{
			name:     "multiple special chars",
			repoName: "my-cool_repo.v2",
			wantErr:  false,
		},
		{
			name:     "single character",
			repoName: "a",
			wantErr:  false,
		},
		{
			name:     "max length 100 chars",
			repoName: "a234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			wantErr:  false,
		},
		{
			name:     "starting with dot",
			repoName: ".github",
			wantErr:  false,
		},
		{
			name:     "starting with underscore",
			repoName: "_private",
			wantErr:  false,
		},
		{
			name:     "only dots",
			repoName: "...",
			wantErr:  false,
		},
		// Invalid repo names
		{
			name:     "empty string",
			repoName: "",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
		{
			name:     "with space",
			repoName: "my repo",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
		{
			name:     "with special characters",
			repoName: "my@repo",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
		{
			name:     "with slash",
			repoName: "my/repo",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
		{
			name:     "with backslash",
			repoName: "my\\repo",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
		{
			name:     "too long (101 chars)",
			repoName: "a2345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
		{
			name:     "with unicode",
			repoName: "my-repo-\u65e5\u672c",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
		{
			name:     "with colon",
			repoName: "my:repo",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
		{
			name:     "with asterisk",
			repoName: "my*repo",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
		{
			name:     "with question mark",
			repoName: "my?repo",
			wantErr:  true,
			errMsg:   "invalid repository name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepoName(tt.repoName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateRepoName(%q) expected error, got nil", tt.repoName)
					return
				}
				if tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("ValidateRepoName(%q) error = %q, want error containing %q", tt.repoName, err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateRepoName(%q) unexpected error: %v", tt.repoName, err)
			}
		})
	}
}

func TestValidateRepository(t *testing.T) {
	tests := []struct {
		name     string
		owner    string
		repoName string
		wantErr  bool
	}{
		{
			name:     "valid owner and repo",
			owner:    "owner",
			repoName: "repo",
			wantErr:  false,
		},
		{
			name:     "invalid owner",
			owner:    "invalid--owner",
			repoName: "repo",
			wantErr:  true,
		},
		{
			name:     "invalid repo",
			owner:    "owner",
			repoName: "invalid/repo",
			wantErr:  true,
		},
		{
			name:     "both invalid",
			owner:    "invalid--owner",
			repoName: "invalid/repo",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepository(tt.owner, tt.repoName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateRepository(%q, %q) expected error, got nil", tt.owner, tt.repoName)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateRepository(%q, %q) unexpected error: %v", tt.owner, tt.repoName, err)
			}
		})
	}
}

func TestValidateWorkflowPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		// Valid paths
		{
			name:    "simple yml",
			path:    ".github/workflows/ci.yml",
			wantErr: false,
		},
		{
			name:    "simple yaml",
			path:    ".github/workflows/ci.yaml",
			wantErr: false,
		},
		{
			name:    "with hyphens",
			path:    ".github/workflows/build-and-test.yml",
			wantErr: false,
		},
		{
			name:    "with underscores",
			path:    ".github/workflows/build_and_test.yml",
			wantErr: false,
		},
		{
			name:    "with dots",
			path:    ".github/workflows/v1.0.0.yml",
			wantErr: false,
		},
		{
			name:    "with numbers",
			path:    ".github/workflows/test123.yml",
			wantErr: false,
		},
		{
			name:    "complex name",
			path:    ".github/workflows/my-app_v1.2.3-build.yaml",
			wantErr: false,
		},
		// Invalid paths
		{
			name:    "empty string",
			path:    "",
			wantErr: true,
			errMsg:  "invalid workflow path",
		},
		{
			name:    "wrong directory",
			path:    ".github/workflow/ci.yml",
			wantErr: true,
			errMsg:  "invalid workflow path",
		},
		{
			name:    "missing .github prefix",
			path:    "workflows/ci.yml",
			wantErr: true,
			errMsg:  "invalid workflow path",
		},
		{
			name:    "wrong extension",
			path:    ".github/workflows/ci.txt",
			wantErr: true,
			errMsg:  "invalid workflow path",
		},
		{
			name:    "no extension",
			path:    ".github/workflows/ci",
			wantErr: true,
			errMsg:  "invalid workflow path",
		},
		{
			name:    "nested directory",
			path:    ".github/workflows/nested/ci.yml",
			wantErr: true,
			errMsg:  "invalid workflow path",
		},
		{
			name:    "with spaces",
			path:    ".github/workflows/my ci.yml",
			wantErr: true,
			errMsg:  "invalid workflow path",
		},
		{
			name:    "absolute path",
			path:    "/home/user/.github/workflows/ci.yml",
			wantErr: true,
			errMsg:  "invalid workflow path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkflowPath(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateWorkflowPath(%q) expected error, got nil", tt.path)
					return
				}
				if tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("ValidateWorkflowPath(%q) error = %q, want error containing %q", tt.path, err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateWorkflowPath(%q) unexpected error: %v", tt.path, err)
			}
		})
	}
}

// containsStr checks if s contains substr (helper function)
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
