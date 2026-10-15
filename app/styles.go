package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Active theme instance
var CurrentTheme Theme

// UI state colors - semantic aliases
var (
	FocusedColor   lipgloss.Color
	UnfocusedColor lipgloss.Color
)

// Pane styles
var (
	FocusedPane   lipgloss.Style
	UnfocusedPane lipgloss.Style
)

// Title styles
var (
	FocusedTitle   lipgloss.Style
	UnfocusedTitle lipgloss.Style
)

// Status icon styles
var (
	SuccessStyle   lipgloss.Style
	FailureStyle   lipgloss.Style
	RunningStyle   lipgloss.Style
	QueuedStyle    lipgloss.Style
	CancelledStyle lipgloss.Style
)

// Selection styles
var (
	SelectedItemFocused   lipgloss.Style
	SelectedItemUnfocused lipgloss.Style
	CursorStyle           lipgloss.Style
	NormalItem            lipgloss.Style
	SelectedItem          lipgloss.Style // Backward compatibility alias
)

// Dialog styles
var (
	ConfirmDialog lipgloss.Style
	HelpPopup     lipgloss.Style
	StatusBar     lipgloss.Style
)

// Log syntax highlighting styles
var (
	LogTimestampStyle lipgloss.Style
	LogGroupStyle     lipgloss.Style
	LogEndGroupStyle  lipgloss.Style
	LogErrorStyle     lipgloss.Style
	LogWarningStyle   lipgloss.Style
	LogNoticeStyle    lipgloss.Style
	LogErrorKeyword   lipgloss.Style
	LogWarningKeyword lipgloss.Style
	LogSuccessKeyword lipgloss.Style
)

func init() {
	// Apply Catppuccin Mocha by default at startup
	ApplyTheme(CatppuccinMocha)
}

// ApplyTheme configures all exported Lipgloss styles to use the given Theme palette.
func ApplyTheme(t Theme) {
	CurrentTheme = t

	FocusedColor = t.FocusedColor
	UnfocusedColor = t.UnfocusedColor

	FocusedPane = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.FocusedColor)

	UnfocusedPane = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.UnfocusedColor)

	FocusedTitle = lipgloss.NewStyle().
		Background(t.FocusedColor).
		Foreground(t.StatusBarBg).
		Bold(true)

	UnfocusedTitle = lipgloss.NewStyle().
		Foreground(t.UnfocusedColor)

	SuccessStyle = lipgloss.NewStyle().Foreground(t.StatusSuccess)
	FailureStyle = lipgloss.NewStyle().Foreground(t.StatusFailure)
	RunningStyle = lipgloss.NewStyle().Foreground(t.StatusRunning)
	QueuedStyle = lipgloss.NewStyle().Foreground(t.StatusQueued)
	CancelledStyle = lipgloss.NewStyle().Foreground(t.StatusCancelled)

	SelectedItemFocused = lipgloss.NewStyle().
		Foreground(t.SelectedItemFg).
		Background(t.SelectedItemBg).
		Bold(true)

	SelectedItemUnfocused = lipgloss.NewStyle().
		Foreground(t.UnfocusedSelectedItemFg).
		Background(t.UnfocusedSelectedItemBg)

	CursorStyle = lipgloss.NewStyle().
		Foreground(t.CursorFg).
		Bold(true)

	NormalItem = lipgloss.NewStyle().
		Foreground(t.NormalItemFg)

	SelectedItem = SelectedItemFocused

	ConfirmDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ConfirmBorder).
		Padding(1, 2)

	HelpPopup = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(t.HelpBorder).
		Padding(1, 2)

	StatusBar = lipgloss.NewStyle().
		Background(t.StatusBarBg).
		Foreground(t.StatusBarFg).
		Padding(0, 1)

	LogTimestampStyle = lipgloss.NewStyle().Foreground(t.LogTimestamp)
	LogGroupStyle = lipgloss.NewStyle().Foreground(t.LogGroup).Bold(true)
	LogEndGroupStyle = lipgloss.NewStyle().Foreground(t.LogEndGroup)
	LogErrorStyle = lipgloss.NewStyle().Foreground(t.LogError).Bold(true)
	LogWarningStyle = lipgloss.NewStyle().Foreground(t.LogWarning)
	LogNoticeStyle = lipgloss.NewStyle().Foreground(t.LogNotice)
	LogErrorKeyword = lipgloss.NewStyle().Foreground(t.LogErrorKeyword)
	LogWarningKeyword = lipgloss.NewStyle().Foreground(t.LogWarningKeyword)
	LogSuccessKeyword = lipgloss.NewStyle().Foreground(t.LogSuccessKeyword)
}

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
