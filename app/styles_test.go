package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// =============================================================================
// StatusIcon Tests
// =============================================================================

func TestStatusIcon_InProgress(t *testing.T) {
	icon := StatusIcon("in_progress", "")
	if !strings.Contains(icon, "●") {
		t.Errorf("StatusIcon(in_progress, \"\") = %q, want to contain ●", icon)
	}
}

func TestStatusIcon_Queued(t *testing.T) {
	icon := StatusIcon("queued", "")
	if !strings.Contains(icon, "○") {
		t.Errorf("StatusIcon(queued, \"\") = %q, want to contain ○", icon)
	}
}

func TestStatusIcon_Success(t *testing.T) {
	icon := StatusIcon("completed", "success")
	if !strings.Contains(icon, "✓") {
		t.Errorf("StatusIcon(completed, success) = %q, want to contain ✓", icon)
	}
}

func TestStatusIcon_Failure(t *testing.T) {
	icon := StatusIcon("completed", "failure")
	if !strings.Contains(icon, "✗") {
		t.Errorf("StatusIcon(completed, failure) = %q, want to contain ✗", icon)
	}
}

func TestStatusIcon_Cancelled(t *testing.T) {
	icon := StatusIcon("completed", "cancelled")
	if !strings.Contains(icon, "⊘") {
		t.Errorf("StatusIcon(completed, cancelled) = %q, want to contain ⊘", icon)
	}
}

func TestStatusIcon_Default(t *testing.T) {
	icon := StatusIcon("unknown", "unknown")
	if icon != " " {
		t.Errorf("StatusIcon(unknown, unknown) = %q, want \" \"", icon)
	}
}

func TestStatusIcon_StatusTakesPrecedence(t *testing.T) {
	// When status is in_progress, conclusion should be ignored
	icon := StatusIcon("in_progress", "failure")
	if !strings.Contains(icon, "●") {
		t.Errorf("StatusIcon(in_progress, failure) = %q, want to contain ● (status takes precedence)", icon)
	}
}

func TestStatusIcon_QueuedIgnoresConclusion(t *testing.T) {
	// When status is queued, conclusion should be ignored
	icon := StatusIcon("queued", "success")
	if !strings.Contains(icon, "○") {
		t.Errorf("StatusIcon(queued, success) = %q, want to contain ○ (queued takes precedence)", icon)
	}
}

func TestStatusIcon_AllStatuses(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		conclusion string
		wantIcon   string
	}{
		{"in_progress", "in_progress", "", "●"},
		{"queued", "queued", "", "○"},
		{"completed_success", "completed", "success", "✓"},
		{"completed_failure", "completed", "failure", "✗"},
		{"completed_cancelled", "completed", "cancelled", "⊘"},
		{"empty_empty", "", "", " "},
		{"completed_skipped", "completed", "skipped", " "},
		{"waiting", "waiting", "", " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := StatusIcon(tt.status, tt.conclusion)
			if !strings.Contains(icon, tt.wantIcon) {
				t.Errorf("StatusIcon(%q, %q) = %q, want to contain %q",
					tt.status, tt.conclusion, icon, tt.wantIcon)
			}
		})
	}
}

// =============================================================================
// RenderItem Tests
// =============================================================================

func TestRenderItem_Selected(t *testing.T) {
	result := RenderItem("test item", true)
	if !strings.Contains(result, "> test item") {
		t.Errorf("RenderItem(\"test item\", true) = %q, want to contain \"> test item\"", result)
	}
}

func TestRenderItem_NotSelected(t *testing.T) {
	result := RenderItem("test item", false)
	if !strings.Contains(result, "  test item") {
		t.Errorf("RenderItem(\"test item\", false) = %q, want to contain \"  test item\"", result)
	}
}

func TestRenderItem_EmptyText(t *testing.T) {
	selectedResult := RenderItem("", true)
	if !strings.Contains(selectedResult, "> ") {
		t.Errorf("RenderItem(\"\", true) = %q, want to contain \"> \"", selectedResult)
	}

	unselectedResult := RenderItem("", false)
	if !strings.Contains(unselectedResult, "  ") {
		t.Errorf("RenderItem(\"\", false) = %q, want to contain \"  \"", unselectedResult)
	}
}

func TestRenderItem_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		selected bool
	}{
		{"unicode", "テスト", true},
		{"emoji", "build ✓", false},
		{"spaces", "item with spaces", true},
		{"numbers", "run #123", false},
		{"symbols", "ci.yml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderItem(tt.text, tt.selected)
			if !strings.Contains(result, tt.text) {
				t.Errorf("RenderItem(%q, %v) = %q, want to contain %q",
					tt.text, tt.selected, result, tt.text)
			}
		})
	}
}

func TestRenderItem_LongText(t *testing.T) {
	longText := "this is a very long text that might be displayed in a list item"
	result := RenderItem(longText, true)
	if !strings.Contains(result, longText) {
		t.Errorf("RenderItem with long text should contain the full text")
	}
}

// =============================================================================
// ScrollPosition Tests
// =============================================================================

func TestScrollPosition_Normal(t *testing.T) {
	// Input is 0-indexed, output is 1-indexed for display
	result := ScrollPosition(0, 10)
	if result != "1/10" {
		t.Errorf("ScrollPosition(0, 10) = %q, want \"1/10\"", result)
	}
}

func TestScrollPosition_FirstOfMany(t *testing.T) {
	result := ScrollPosition(0, 100)
	if result != "1/100" {
		t.Errorf("ScrollPosition(0, 100) = %q, want \"1/100\"", result)
	}
}

func TestScrollPosition_LastOfMany(t *testing.T) {
	result := ScrollPosition(99, 100)
	if result != "100/100" {
		t.Errorf("ScrollPosition(99, 100) = %q, want \"100/100\"", result)
	}
}

func TestScrollPosition_SingleItem(t *testing.T) {
	result := ScrollPosition(0, 1)
	if result != "1/1" {
		t.Errorf("ScrollPosition(0, 1) = %q, want \"1/1\"", result)
	}
}

func TestScrollPosition_ZeroTotal(t *testing.T) {
	result := ScrollPosition(0, 0)
	if result != "0/0" {
		t.Errorf("ScrollPosition(0, 0) = %q, want \"0/0\"", result)
	}
}

func TestScrollPosition_NegativeTotal(t *testing.T) {
	result := ScrollPosition(0, -1)
	if result != "0/0" {
		t.Errorf("ScrollPosition(0, -1) = %q, want \"0/0\"", result)
	}
}

func TestScrollPosition_MiddlePosition(t *testing.T) {
	result := ScrollPosition(4, 10)
	if result != "5/10" {
		t.Errorf("ScrollPosition(4, 10) = %q, want \"5/10\"", result)
	}
}

func TestScrollPosition_LargeNumbers(t *testing.T) {
	result := ScrollPosition(998, 1000)
	if result != "999/1000" {
		t.Errorf("ScrollPosition(998, 1000) = %q, want \"999/1000\"", result)
	}
}

// =============================================================================
// Style Variable Tests (Ensure styles are properly defined)
// =============================================================================

func TestStyleVariables_NotNil(t *testing.T) {
	// These tests ensure that the style variables are properly initialized
	// by attempting to use them

	tests := []struct {
		name  string
		style func() string
	}{
		{"FocusedPane", func() string { return FocusedPane.Render("test") }},
		{"UnfocusedPane", func() string { return UnfocusedPane.Render("test") }},
		{"FocusedTitle", func() string { return FocusedTitle.Render("test") }},
		{"UnfocusedTitle", func() string { return UnfocusedTitle.Render("test") }},
		{"SuccessStyle", func() string { return SuccessStyle.Render("test") }},
		{"FailureStyle", func() string { return FailureStyle.Render("test") }},
		{"RunningStyle", func() string { return RunningStyle.Render("test") }},
		{"QueuedStyle", func() string { return QueuedStyle.Render("test") }},
		{"CancelledStyle", func() string { return CancelledStyle.Render("test") }},
		{"SelectedItem", func() string { return SelectedItem.Render("test") }},
		{"NormalItem", func() string { return NormalItem.Render("test") }},
		{"ConfirmDialog", func() string { return ConfirmDialog.Render("test") }},
		{"HelpPopup", func() string { return HelpPopup.Render("test") }},
		{"StatusBar", func() string { return StatusBar.Render("test") }},
		// Log syntax highlighting styles
		{"LogTimestampStyle", func() string { return LogTimestampStyle.Render("test") }},
		{"LogGroupStyle", func() string { return LogGroupStyle.Render("test") }},
		{"LogEndGroupStyle", func() string { return LogEndGroupStyle.Render("test") }},
		{"LogErrorStyle", func() string { return LogErrorStyle.Render("test") }},
		{"LogWarningStyle", func() string { return LogWarningStyle.Render("test") }},
		{"LogNoticeStyle", func() string { return LogNoticeStyle.Render("test") }},
		{"LogErrorKeyword", func() string { return LogErrorKeyword.Render("test") }},
		{"LogWarningKeyword", func() string { return LogWarningKeyword.Render("test") }},
		{"LogSuccessKeyword", func() string { return LogSuccessKeyword.Render("test") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic and should return non-empty string
			result := tt.style()
			if result == "" {
				t.Errorf("%s.Render(\"test\") returned empty string", tt.name)
			}
		})
	}
}

func TestColorVariables(t *testing.T) {
	// Verify color values
	if FocusedColor != lipgloss.Color("#00FF00") {
		t.Errorf("FocusedColor = %v, want #00FF00", FocusedColor)
	}
	if UnfocusedColor != lipgloss.Color("#666666") {
		t.Errorf("UnfocusedColor = %v, want #666666", UnfocusedColor)
	}
}
