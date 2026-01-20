package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors (lazygit-style)
	FocusedColor   = lipgloss.Color("#00FF00")
	UnfocusedColor = lipgloss.Color("#666666")

	// Pane styles - use thin border for compact UI
	FocusedPane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(FocusedColor)

	UnfocusedPane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(UnfocusedColor)

	// Title styles - lazydocker style inverted title for focused panel
	FocusedTitle = lipgloss.NewStyle().
			Background(FocusedColor).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)

	UnfocusedTitle = lipgloss.NewStyle().
			Foreground(UnfocusedColor)

	// Status icons
	SuccessStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	FailureStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	RunningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	QueuedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	CancelledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800"))

	// Selection - lazydocker style: bright selection for focused, dim for unfocused
	SelectedItemFocused = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#0066CC")).
				Bold(true)

	SelectedItemUnfocused = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC")).
				Background(lipgloss.Color("#444444"))

	// Cursor style for selected item
	CursorStyle = lipgloss.NewStyle().
			Foreground(FocusedColor).
			Bold(true)

	NormalItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))

	// Keep backward compatibility
	SelectedItem = SelectedItemFocused

	// Dialogs
	ConfirmDialog = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF8800")).
			Padding(1, 2)

	HelpPopup = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#00FFFF")).
			Padding(1, 2)

	StatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	// Log syntax highlighting styles
	LogTimestampStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	LogGroupStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	LogEndGroupStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#006600"))
	LogErrorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	LogWarningStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800"))
	LogNoticeStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	LogErrorKeyword    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	LogWarningKeyword  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))
	LogSuccessKeyword  = lipgloss.NewStyle().Foreground(lipgloss.Color("#66FF66"))
)

// StatusIcon returns icon for status
func StatusIcon(status, conclusion string) string {
	switch {
	case status == "in_progress":
		return RunningStyle.Render("●")
	case status == "queued":
		return QueuedStyle.Render("○")
	case conclusion == "success":
		return SuccessStyle.Render("✓")
	case conclusion == "failure":
		return FailureStyle.Render("✗")
	case conclusion == "cancelled":
		return CancelledStyle.Render("⊘")
	default:
		return " "
	}
}

// RenderItem renders list item with selection state
func RenderItem(text string, selected bool) string {
	if selected {
		return SelectedItem.Render("> " + text)
	}
	return NormalItem.Render("  " + text)
}

// ScrollPosition renders scroll position in "1/10" format (1-indexed for display).
func ScrollPosition(current, total int) string {
	if total <= 0 {
		return "0/0"
	}
	return fmt.Sprintf("%d/%d", current+1, total)
}
