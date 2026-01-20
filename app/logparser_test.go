package app

import (
	"testing"
)

func TestParseLogs_SingleStep(t *testing.T) {
	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Run actions/checkout@v4
2024-01-15T10:00:01.000Z with: repository: test/repo
2024-01-15T10:00:02.000Z ##[endgroup]`

	parsed := ParseLogs(rawLogs)

	if len(parsed.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(parsed.Steps))
	}

	step := parsed.Steps[0]
	if step.Name != "Run actions/checkout@v4" {
		t.Errorf("expected step name 'Run actions/checkout@v4', got '%s'", step.Name)
	}

	if len(step.Lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(step.Lines))
	}
}

func TestParseLogs_MultipleSteps(t *testing.T) {
	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Set up job
2024-01-15T10:00:01.000Z Setting up runner...
2024-01-15T10:00:02.000Z ##[endgroup]
2024-01-15T10:00:03.000Z ##[group]Run actions/checkout@v4
2024-01-15T10:00:04.000Z Checking out repository
2024-01-15T10:00:05.000Z ##[endgroup]
2024-01-15T10:00:06.000Z ##[group]Run go test -v ./...
2024-01-15T10:00:07.000Z === RUN TestFoo
2024-01-15T10:00:08.000Z --- PASS: TestFoo (0.01s)
2024-01-15T10:00:09.000Z ##[endgroup]`

	parsed := ParseLogs(rawLogs)

	if len(parsed.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(parsed.Steps))
	}

	expectedNames := []string{"Set up job", "Run actions/checkout@v4", "Run go test -v ./..."}
	for i, expected := range expectedNames {
		if parsed.Steps[i].Name != expected {
			t.Errorf("step %d: expected name '%s', got '%s'", i, expected, parsed.Steps[i].Name)
		}
	}
}

func TestParseLogs_NoSteps(t *testing.T) {
	rawLogs := `2024-01-15T10:00:00.000Z Some plain log output
2024-01-15T10:00:01.000Z Another line without step markers`

	parsed := ParseLogs(rawLogs)

	if len(parsed.Steps) != 0 {
		t.Errorf("expected 0 steps for logs without markers, got %d", len(parsed.Steps))
	}

	if len(parsed.AllLines) != 2 {
		t.Errorf("expected 2 lines in AllLines, got %d", len(parsed.AllLines))
	}
}

func TestParseLogs_UnclosedGroup(t *testing.T) {
	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Running step
2024-01-15T10:00:01.000Z Still running...
2024-01-15T10:00:02.000Z More output`

	parsed := ParseLogs(rawLogs)

	if len(parsed.Steps) != 1 {
		t.Fatalf("expected 1 step for unclosed group, got %d", len(parsed.Steps))
	}

	if len(parsed.Steps[0].Lines) != 3 {
		t.Errorf("expected 3 lines for unclosed group, got %d", len(parsed.Steps[0].Lines))
	}
}

func TestParsedLogs_GetStepLogs_AllLogs(t *testing.T) {
	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Step 1
2024-01-15T10:00:01.000Z Line 1
2024-01-15T10:00:02.000Z ##[endgroup]
2024-01-15T10:00:03.000Z ##[group]Step 2
2024-01-15T10:00:04.000Z Line 2
2024-01-15T10:00:05.000Z ##[endgroup]`

	parsed := ParseLogs(rawLogs)

	allLogs := parsed.GetStepLogs(-1)
	if allLogs != rawLogs {
		t.Errorf("GetStepLogs(-1) should return raw logs")
	}
}

func TestParsedLogs_GetStepLogs_SpecificStep(t *testing.T) {
	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Step 1
2024-01-15T10:00:01.000Z Line 1
2024-01-15T10:00:02.000Z ##[endgroup]
2024-01-15T10:00:03.000Z ##[group]Step 2
2024-01-15T10:00:04.000Z Line 2
2024-01-15T10:00:05.000Z ##[endgroup]`

	parsed := ParseLogs(rawLogs)

	step1Logs := parsed.GetStepLogs(0)
	if step1Logs == "" {
		t.Error("GetStepLogs(0) should return step 1 logs")
	}

	step2Logs := parsed.GetStepLogs(1)
	if step2Logs == "" {
		t.Error("GetStepLogs(1) should return step 2 logs")
	}

	// Out of bounds
	outOfBounds := parsed.GetStepLogs(99)
	if outOfBounds != "" {
		t.Error("GetStepLogs with out of bounds index should return empty string")
	}
}

func TestParsedLogs_GetStepLogs_EmptyLogs(t *testing.T) {
	parsed := ParseLogs("")

	if parsed.GetStepLogs(-1) != "" {
		t.Error("GetStepLogs(-1) on empty logs should return empty string")
	}
}

func TestParseLogs_TimestampExtraction(t *testing.T) {
	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Test Step
2024-01-15T10:00:01.123456789Z Some output with nanoseconds
2024-01-15T10:00:02.000Z ##[endgroup]`

	parsed := ParseLogs(rawLogs)

	if len(parsed.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(parsed.Steps))
	}

	// Verify lines are stored (including timestamps)
	if len(parsed.Steps[0].Lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(parsed.Steps[0].Lines))
	}
}

func TestFormatLogLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with timestamp",
			input:    "2024-01-15T10:00:00.000Z Some log message",
			expected: "10:00:00 Some log message",
		},
		{
			name:     "with nanoseconds",
			input:    "2024-01-15T10:30:45.123456789Z Another message",
			expected: "10:30:45 Another message",
		},
		{
			name:     "no timestamp",
			input:    "Plain message without timestamp",
			expected: "Plain message without timestamp",
		},
		{
			name:     "empty line",
			input:    "",
			expected: "",
		},
		{
			name:     "group marker",
			input:    "2024-01-15T10:00:00.000Z ##[group]Step name",
			expected: "10:00:00 ##[group]Step name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatLogLine(tt.input)
			if result != tt.expected {
				t.Errorf("FormatLogLine(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
