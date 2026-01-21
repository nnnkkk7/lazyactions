package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - semantic color names for consistent theming
var (
	// Primary colors
	ColorGreen      = lipgloss.Color("#00FF00") // Success, focused, running
	ColorRed        = lipgloss.Color("#FF0000") // Error, failure
	ColorYellow     = lipgloss.Color("#FFFF00") // Running, in-progress
	ColorCyan       = lipgloss.Color("#00FFFF") // Info, timestamp, notice
	ColorOrange     = lipgloss.Color("#FF8800") // Warning, cancelled
	ColorBlack      = lipgloss.Color("#000000") // Text on bright backgrounds
	ColorWhite      = lipgloss.Color("#FFFFFF") // Bright text
	ColorDarkGray   = lipgloss.Color("#333333") // Status bar background
	ColorMediumGray = lipgloss.Color("#666666") // Unfocused elements
	ColorLightGray  = lipgloss.Color("#888888") // Queued, dim text
	ColorSilver     = lipgloss.Color("#AAAAAA") // Normal text
	ColorPaleGray   = lipgloss.Color("#CCCCCC") // Unfocused selected text

	// Accent colors
	ColorBlue         = lipgloss.Color("#0066CC") // Selection background
	ColorDarkBlue     = lipgloss.Color("#444444") // Unfocused selection background
	ColorDarkGreen    = lipgloss.Color("#006600") // End group marker
	ColorLightRed     = lipgloss.Color("#FF6666") // Error keyword highlight
	ColorLightOrange  = lipgloss.Color("#FFAA00") // Warning keyword highlight
	ColorLightGreen   = lipgloss.Color("#66FF66") // Success keyword highlight
)

// UI state colors - semantic aliases
var (
	FocusedColor   = ColorGreen
	UnfocusedColor = ColorMediumGray
)

// Pane styles - use thin border for compact UI
var (
	FocusedPane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(FocusedColor)

	UnfocusedPane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(UnfocusedColor)
)

// Title styles - lazydocker style inverted title for focused panel
var (
	FocusedTitle = lipgloss.NewStyle().
			Background(FocusedColor).
			Foreground(ColorBlack).
			Bold(true)

	UnfocusedTitle = lipgloss.NewStyle().
			Foreground(UnfocusedColor)
)

// Status icon styles
var (
	SuccessStyle   = lipgloss.NewStyle().Foreground(ColorGreen)
	FailureStyle   = lipgloss.NewStyle().Foreground(ColorRed)
	RunningStyle   = lipgloss.NewStyle().Foreground(ColorYellow)
	QueuedStyle    = lipgloss.NewStyle().Foreground(ColorLightGray)
	CancelledStyle = lipgloss.NewStyle().Foreground(ColorOrange)
)

// Selection styles - lazydocker style: bright selection for focused, dim for unfocused
var (
	SelectedItemFocused = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorBlue).
				Bold(true)

	SelectedItemUnfocused = lipgloss.NewStyle().
				Foreground(ColorPaleGray).
				Background(ColorDarkBlue)

	// Cursor style for selected item
	CursorStyle = lipgloss.NewStyle().
			Foreground(FocusedColor).
			Bold(true)

	NormalItem = lipgloss.NewStyle().
			Foreground(ColorSilver)

	// Keep backward compatibility
	SelectedItem = SelectedItemFocused
)

// Dialog styles
var (
	ConfirmDialog = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorOrange).
			Padding(1, 2)

	HelpPopup = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2)

	StatusBar = lipgloss.NewStyle().
			Background(ColorDarkGray).
			Padding(0, 1)
)

// Log syntax highlighting styles
var (
	LogTimestampStyle = lipgloss.NewStyle().Foreground(ColorCyan)
	LogGroupStyle     = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	LogEndGroupStyle  = lipgloss.NewStyle().Foreground(ColorDarkGreen)
	LogErrorStyle     = lipgloss.NewStyle().Foreground(ColorRed).Bold(true)
	LogWarningStyle   = lipgloss.NewStyle().Foreground(ColorOrange)
	LogNoticeStyle    = lipgloss.NewStyle().Foreground(ColorCyan)
	LogErrorKeyword   = lipgloss.NewStyle().Foreground(ColorLightRed)
	LogWarningKeyword = lipgloss.NewStyle().Foreground(ColorLightOrange)
	LogSuccessKeyword = lipgloss.NewStyle().Foreground(ColorLightGreen)
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
