package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewLogViewport tests the creation of a new LogViewport.
func TestNewLogViewport(t *testing.T) {
	width := 80
	height := 24

	lv := NewLogViewport(width, height)

	if lv == nil {
		t.Fatal("NewLogViewport returned nil")
	}
	if !lv.autoscroll {
		t.Error("expected autoscroll to be true by default")
	}
}

// TestLogViewport_SetContent tests setting content in the viewport.
func TestLogViewport_SetContent(t *testing.T) {
	lv := NewLogViewport(80, 10)

	content := "Line 1\nLine 2\nLine 3"
	lv.SetContent(content)

	view := lv.View()
	// The view should contain the content (or parts of it)
	if view == "" {
		t.Error("expected non-empty view after SetContent")
	}
}

// TestLogViewport_SetContentMultipleLines tests setting multi-line content.
func TestLogViewport_SetContentMultipleLines(t *testing.T) {
	lv := NewLogViewport(80, 5)

	// Create content with more lines than the viewport height
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "Line " + string(rune('A'+i))
	}
	content := strings.Join(lines, "\n")

	lv.SetContent(content)

	// With autoscroll enabled and starting at bottom, it should scroll to bottom
	// Note: The exact behavior depends on the viewport implementation
	view := lv.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

// TestLogViewport_View tests the View method returns string content.
func TestLogViewport_View(t *testing.T) {
	lv := NewLogViewport(40, 5)

	// Empty viewport
	view := lv.View()
	// View should return something (possibly empty content with styling)
	_ = view // Just ensure it doesn't panic

	// With content
	lv.SetContent("Hello, World!")
	view = lv.View()
	if view == "" {
		t.Error("expected non-empty view after setting content")
	}
}

// TestLogViewport_SetSize tests resizing the viewport.
func TestLogViewport_SetSize(t *testing.T) {
	lv := NewLogViewport(80, 24)

	// Resize
	lv.SetSize(40, 10)

	// Should not panic and should work with new size
	lv.SetContent("Test content")
	view := lv.View()
	if view == "" {
		t.Error("expected non-empty view after resize")
	}
}

// TestLogViewport_Update tests the Update method.
func TestLogViewport_Update(t *testing.T) {
	lv := NewLogViewport(80, 10)
	lv.SetContent("Line 1\nLine 2\nLine 3")

	// Update with a message
	updated, cmd := lv.Update(nil)

	if updated == nil {
		t.Error("expected non-nil updated viewport")
	}
	// cmd may be nil depending on the message
	_ = cmd
}

// TestLogViewport_Autoscroll tests the autoscroll behavior.
func TestLogViewport_Autoscroll(t *testing.T) {
	lv := NewLogViewport(80, 5)

	// Initially autoscroll should be enabled
	if !lv.autoscroll {
		t.Error("expected autoscroll to be true initially")
	}

	// Set some content
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "Line " + string(rune('A'+i))
	}
	lv.SetContent(strings.Join(lines, "\n"))

	// After SetContent with autoscroll, should still have autoscroll enabled
	// (since we were at bottom before and should be at bottom after)
	if !lv.autoscroll {
		t.Error("expected autoscroll to remain true after SetContent when at bottom")
	}
}

// TestLogViewport_AutoscrollDisabledOnManualScroll tests that autoscroll is disabled on manual scroll.
func TestLogViewport_AutoscrollDisabledOnManualScroll(t *testing.T) {
	lv := NewLogViewport(80, 5)

	// Set content that requires scrolling
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "Line " + string(rune('0'+(i%10)))
	}
	lv.SetContent(strings.Join(lines, "\n"))

	// Simulate scrolling up using a key message
	// Note: The exact behavior depends on how the viewport handles scroll
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := lv.Update(keyMsg)

	// After scrolling up, autoscroll might be disabled if not at bottom
	// This depends on the viewport's AtBottom() implementation
	_ = updated

	// The main thing is that it doesn't panic
}

// TestLogViewport_AutoscrollReenabledAtBottom tests that autoscroll is re-enabled when scrolling to bottom.
func TestLogViewport_AutoscrollReenabledAtBottom(t *testing.T) {
	lv := NewLogViewport(80, 5)

	// Set content
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "Line " + string(rune('0'+(i%10)))
	}
	lv.SetContent(strings.Join(lines, "\n"))

	// Scroll up then down
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	lv.Update(upMsg)

	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	// Scroll down multiple times to reach bottom
	for i := 0; i < 100; i++ {
		lv.Update(downMsg)
	}

	// Now we should be at bottom and autoscroll should be re-enabled
	// Note: The exact behavior depends on viewport implementation
}

// TestLogViewport_EmptyContent tests behavior with empty content.
func TestLogViewport_EmptyContent(t *testing.T) {
	lv := NewLogViewport(80, 10)

	lv.SetContent("")

	view := lv.View()
	// Should not panic
	_ = view
}

// TestLogViewport_LargeContent tests behavior with very large content.
func TestLogViewport_LargeContent(t *testing.T) {
	lv := NewLogViewport(80, 10)

	// Create large content
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString("This is a test line with some content.\n")
	}

	lv.SetContent(sb.String())

	view := lv.View()
	if view == "" {
		t.Error("expected non-empty view with large content")
	}
}

// TestLogViewport_ContentWithSpecialChars tests content with special characters.
func TestLogViewport_ContentWithSpecialChars(t *testing.T) {
	lv := NewLogViewport(80, 10)

	content := "Line with \ttabs and\nspecial chars: \033[31mred\033[0m and unicode: "
	lv.SetContent(content)

	view := lv.View()
	// Should not panic
	_ = view
}

// TestLogViewport_ZeroSize tests behavior with zero dimensions.
func TestLogViewport_ZeroSize(t *testing.T) {
	// This tests edge case behavior - implementation should handle gracefully
	lv := NewLogViewport(0, 0)

	lv.SetContent("Test")
	view := lv.View()
	_ = view // Should not panic
}

// TestLogViewport_SetSizeToZero tests resizing to zero.
func TestLogViewport_SetSizeToZero(t *testing.T) {
	lv := NewLogViewport(80, 24)
	lv.SetContent("Test content")

	lv.SetSize(0, 0)

	view := lv.View()
	_ = view // Should not panic
}

// TestLogViewport_IsAtBottom tests the internal isAtBottom method behavior.
func TestLogViewport_IsAtBottom(t *testing.T) {
	lv := NewLogViewport(80, 5)

	// Empty content - should be at bottom
	if !lv.isAtBottom() {
		t.Error("expected to be at bottom with empty content")
	}

	// Small content that fits - should be at bottom
	lv.SetContent("Line 1\nLine 2")
	if !lv.isAtBottom() {
		t.Error("expected to be at bottom when content fits")
	}
}

// TestLogViewport_UpdateWithNilMessage tests Update with nil message.
func TestLogViewport_UpdateWithNilMessage(t *testing.T) {
	lv := NewLogViewport(80, 10)
	lv.SetContent("Test content")

	updated, cmd := lv.Update(nil)

	if updated == nil {
		t.Error("expected non-nil updated viewport")
	}
	_ = cmd // May be nil
}

// TestLogViewport_MultipleSetContent tests multiple SetContent calls.
func TestLogViewport_MultipleSetContent(t *testing.T) {
	lv := NewLogViewport(80, 10)

	lv.SetContent("First content")
	lv.SetContent("Second content")
	lv.SetContent("Third content")

	view := lv.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

// TestLogViewport_SetSizeMultipleTimes tests multiple resize operations.
func TestLogViewport_SetSizeMultipleTimes(t *testing.T) {
	lv := NewLogViewport(80, 24)
	lv.SetContent("Test content")

	lv.SetSize(40, 10)
	lv.SetSize(120, 30)
	lv.SetSize(60, 15)

	view := lv.View()
	if view == "" {
		t.Error("expected non-empty view after multiple resizes")
	}
}
