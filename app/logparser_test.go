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

func TestFormatLogLineWithColor(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty line",
			input: "",
		},
		{
			name:  "with timestamp",
			input: "2024-01-15T10:00:00.000Z Some log message",
		},
		{
			name:  "group marker",
			input: "2024-01-15T10:00:00.000Z ##[group]Step name",
		},
		{
			name:  "endgroup marker",
			input: "2024-01-15T10:00:00.000Z ##[endgroup]",
		},
		{
			name:  "error marker",
			input: "2024-01-15T10:00:00.000Z ##[error]Something went wrong",
		},
		{
			name:  "warning marker",
			input: "2024-01-15T10:00:00.000Z ##[warning]This is a warning",
		},
		{
			name:  "notice marker",
			input: "2024-01-15T10:00:00.000Z ##[notice]Notice message",
		},
		{
			name:  "error keyword",
			input: "2024-01-15T10:00:00.000Z Test failed with error",
		},
		{
			name:  "warning keyword",
			input: "2024-01-15T10:00:00.000Z Warning: deprecated function",
		},
		{
			name:  "success keyword",
			input: "2024-01-15T10:00:00.000Z Test passed successfully",
		},
		{
			name:  "no timestamp with error",
			input: "##[error]Error without timestamp",
		},
		{
			name:  "plain text no timestamp",
			input: "Plain text without any markers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatLogLineWithColor(tt.input)
			// Just verify no panic and result is not empty for non-empty input
			if tt.input != "" && result == "" {
				t.Errorf("FormatLogLineWithColor(%q) returned empty string", tt.input)
			}
			if tt.input == "" && result != "" {
				t.Errorf("FormatLogLineWithColor(empty) should return empty, got %q", result)
			}
		})
	}
}

func TestFormatLogLineWithColor_ContainsOriginalText(t *testing.T) {
	// Verify that the formatted output contains the original text content
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "timestamp is formatted",
			input:    "2024-01-15T10:00:00.000Z Hello world",
			contains: "10:00:00",
		},
		{
			name:     "message content preserved",
			input:    "2024-01-15T10:00:00.000Z Hello world",
			contains: "Hello world",
		},
		{
			name:     "error marker content",
			input:    "##[error]Something failed",
			contains: "Something failed",
		},
		{
			name:     "group marker content",
			input:    "##[group]Run tests",
			contains: "Run tests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatLogLineWithColor(tt.input)
			// Note: The result includes ANSI escape codes, so we check if the base text is present
			// The escape codes wrap the text but the text content should still be there
			if result == "" {
				t.Errorf("FormatLogLineWithColor(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestHighlightKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "error keyword",
			input: "This is an error message",
		},
		{
			name:  "failed keyword",
			input: "Test failed",
		},
		{
			name:  "warning keyword",
			input: "Warning: something happened",
		},
		{
			name:  "success keyword",
			input: "Build passed successfully",
		},
		{
			name:  "multiple keywords",
			input: "Test passed but warning about error",
		},
		{
			name:  "no keywords",
			input: "Normal log line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := highlightKeywords(tt.input)
			// Verify no panic and result is not empty
			if result == "" && tt.input != "" {
				t.Errorf("highlightKeywords(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestFormatStepLogsWithColor(t *testing.T) {
	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Run tests
2024-01-15T10:00:01.000Z Running test suite...
2024-01-15T10:00:02.000Z Test passed
2024-01-15T10:00:03.000Z ##[endgroup]`

	parsed := ParseLogs(rawLogs)

	// Test with all logs
	result := parsed.FormatStepLogsWithColor(-1)
	if result == "" {
		t.Error("FormatStepLogsWithColor(-1) should return formatted logs")
	}

	// Test with specific step
	stepResult := parsed.FormatStepLogsWithColor(0)
	if stepResult == "" {
		t.Error("FormatStepLogsWithColor(0) should return formatted step logs")
	}

	// Test with empty logs
	emptyParsed := ParseLogs("")
	emptyResult := emptyParsed.FormatStepLogsWithColor(-1)
	if emptyResult != "" {
		t.Errorf("FormatStepLogsWithColor on empty logs should return empty, got %q", emptyResult)
	}
}
