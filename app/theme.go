package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color palette used throughout the application UI.
type Theme struct {
	Name string

	// UI Focus & Borders
	FocusedColor   lipgloss.Color
	UnfocusedColor lipgloss.Color

	// Selection & Lists
	SelectedItemFg          lipgloss.Color
	SelectedItemBg          lipgloss.Color
	UnfocusedSelectedItemFg lipgloss.Color
	UnfocusedSelectedItemBg lipgloss.Color
	CursorFg                lipgloss.Color
	NormalItemFg            lipgloss.Color

	// Status Indicators
	StatusSuccess   lipgloss.Color
	StatusFailure   lipgloss.Color
	StatusRunning   lipgloss.Color
	StatusQueued    lipgloss.Color
	StatusCancelled lipgloss.Color

	// Status Bar & Popups
	StatusBarBg   lipgloss.Color
	StatusBarFg   lipgloss.Color
	ConfirmBorder lipgloss.Color
	HelpBorder    lipgloss.Color

	// Logs
	LogTimestamp      lipgloss.Color
	LogGroup          lipgloss.Color
	LogEndGroup       lipgloss.Color
	LogError          lipgloss.Color
	LogWarning        lipgloss.Color
	LogNotice         lipgloss.Color
	LogErrorKeyword   lipgloss.Color
	LogWarningKeyword lipgloss.Color
	LogSuccessKeyword lipgloss.Color
}

// Built-in theme presets
var (
	// CatppuccinMocha is the default dark theme (Catppuccin Mocha palette)
	CatppuccinMocha = Theme{
		Name:                    "catppuccin-mocha",
		FocusedColor:            lipgloss.Color("#89b4fa"), // Blue
		UnfocusedColor:          lipgloss.Color("#585b70"), // Surface2
		SelectedItemFg:          lipgloss.Color("#cdd6f4"), // Text
		SelectedItemBg:          lipgloss.Color("#313244"), // Surface0
		UnfocusedSelectedItemFg: lipgloss.Color("#a6adc8"), // Subtext0
		UnfocusedSelectedItemBg: lipgloss.Color("#181825"), // Mantle
		CursorFg:                lipgloss.Color("#89b4fa"), // Blue
		NormalItemFg:            lipgloss.Color("#bac2de"), // Subtext1
		StatusSuccess:           lipgloss.Color("#a6e3a1"), // Green
		StatusFailure:           lipgloss.Color("#f38ba8"), // Red
		StatusRunning:           lipgloss.Color("#f9e2af"), // Yellow
		StatusQueued:            lipgloss.Color("#6c7086"), // Overlay0
		StatusCancelled:         lipgloss.Color("#fab387"), // Peach
		StatusBarBg:             lipgloss.Color("#181825"), // Mantle
		StatusBarFg:             lipgloss.Color("#cdd6f4"), // Text
		ConfirmBorder:           lipgloss.Color("#fab387"), // Peach
		HelpBorder:              lipgloss.Color("#89dceb"), // Sky
		LogTimestamp:            lipgloss.Color("#89dceb"), // Sky
		LogGroup:                lipgloss.Color("#a6e3a1"), // Green
		LogEndGroup:             lipgloss.Color("#94e2d5"), // Teal
		LogError:                lipgloss.Color("#f38ba8"), // Red
		LogWarning:              lipgloss.Color("#fab387"), // Peach
		LogNotice:               lipgloss.Color("#89b4fa"), // Blue
		LogErrorKeyword:         lipgloss.Color("#eba0ac"), // Maroon
		LogWarningKeyword:       lipgloss.Color("#f9e2af"), // Yellow
		LogSuccessKeyword:       lipgloss.Color("#a6e3a1"), // Green
	}

	// CatppuccinMacchiato (Catppuccin Macchiato palette)
	CatppuccinMacchiato = Theme{
		Name:                    "catppuccin-macchiato",
		FocusedColor:            lipgloss.Color("#8aadf4"), // Blue
		UnfocusedColor:          lipgloss.Color("#5b6078"), // Surface2
		SelectedItemFg:          lipgloss.Color("#cad3f5"), // Text
		SelectedItemBg:          lipgloss.Color("#363a4f"), // Surface0
		UnfocusedSelectedItemFg: lipgloss.Color("#a5adcb"), // Subtext0
		UnfocusedSelectedItemBg: lipgloss.Color("#1e2030"), // Mantle
		CursorFg:                lipgloss.Color("#8aadf4"), // Blue
		NormalItemFg:            lipgloss.Color("#b8c0e0"), // Subtext1
		StatusSuccess:           lipgloss.Color("#a6da95"), // Green
		StatusFailure:           lipgloss.Color("#ed8796"), // Red
		StatusRunning:           lipgloss.Color("#eed49f"), // Yellow
		StatusQueued:            lipgloss.Color("#6e738d"), // Overlay0
		StatusCancelled:         lipgloss.Color("#f5a97f"), // Peach
		StatusBarBg:             lipgloss.Color("#1e2030"), // Mantle
		StatusBarFg:             lipgloss.Color("#cad3f5"), // Text
		ConfirmBorder:           lipgloss.Color("#f5a97f"), // Peach
		HelpBorder:              lipgloss.Color("#91d7e3"), // Sky
		LogTimestamp:            lipgloss.Color("#91d7e3"), // Sky
		LogGroup:                lipgloss.Color("#a6da95"), // Green
		LogEndGroup:             lipgloss.Color("#8bd5ca"), // Teal
		LogError:                lipgloss.Color("#ed8796"), // Red
		LogWarning:              lipgloss.Color("#f5a97f"), // Peach
		LogNotice:               lipgloss.Color("#8aadf4"), // Blue
		LogErrorKeyword:         lipgloss.Color("#ee99a0"), // Maroon
		LogWarningKeyword:       lipgloss.Color("#eed49f"), // Yellow
		LogSuccessKeyword:       lipgloss.Color("#a6da95"), // Green
	}

	// CatppuccinFrappe (Catppuccin Frappé palette)
	CatppuccinFrappe = Theme{
		Name:                    "catppuccin-frappe",
		FocusedColor:            lipgloss.Color("#8caaee"), // Blue
		UnfocusedColor:          lipgloss.Color("#626880"), // Surface2
		SelectedItemFg:          lipgloss.Color("#c6d0f5"), // Text
		SelectedItemBg:          lipgloss.Color("#414559"), // Surface0
		UnfocusedSelectedItemFg: lipgloss.Color("#a5adce"), // Subtext0
		UnfocusedSelectedItemBg: lipgloss.Color("#292c3c"), // Mantle
		CursorFg:                lipgloss.Color("#8caaee"), // Blue
		NormalItemFg:            lipgloss.Color("#b5bfe2"), // Subtext1
		StatusSuccess:           lipgloss.Color("#a6d189"), // Green
		StatusFailure:           lipgloss.Color("#e78284"), // Red
		StatusRunning:           lipgloss.Color("#e5c890"), // Yellow
		StatusQueued:            lipgloss.Color("#737994"), // Overlay0
		StatusCancelled:         lipgloss.Color("#ef9f76"), // Peach
		StatusBarBg:             lipgloss.Color("#292c3c"), // Mantle
		StatusBarFg:             lipgloss.Color("#c6d0f5"), // Text
		ConfirmBorder:           lipgloss.Color("#ef9f76"), // Peach
		HelpBorder:              lipgloss.Color("#99d1db"), // Sky
		LogTimestamp:            lipgloss.Color("#99d1db"), // Sky
		LogGroup:                lipgloss.Color("#a6d189"), // Green
		LogEndGroup:             lipgloss.Color("#81c8be"), // Teal
		LogError:                lipgloss.Color("#e78284"), // Red
		LogWarning:              lipgloss.Color("#ef9f76"), // Peach
		LogNotice:               lipgloss.Color("#8caaee"), // Blue
		LogErrorKeyword:         lipgloss.Color("#ea999c"), // Maroon
		LogWarningKeyword:       lipgloss.Color("#e5c890"), // Yellow
		LogSuccessKeyword:       lipgloss.Color("#a6d189"), // Green
	}

	// CatppuccinLatte (Catppuccin Latte light palette)
	CatppuccinLatte = Theme{
		Name:                    "catppuccin-latte",
		FocusedColor:            lipgloss.Color("#1e66f5"), // Blue
		UnfocusedColor:          lipgloss.Color("#acb0be"), // Surface2
		SelectedItemFg:          lipgloss.Color("#4c4f69"), // Text
		SelectedItemBg:          lipgloss.Color("#ccd0da"), // Surface0
		UnfocusedSelectedItemFg: lipgloss.Color("#6c6f85"), // Subtext0
		UnfocusedSelectedItemBg: lipgloss.Color("#e6e9ef"), // Mantle
		CursorFg:                lipgloss.Color("#1e66f5"), // Blue
		NormalItemFg:            lipgloss.Color("#5c5f77"), // Subtext1
		StatusSuccess:           lipgloss.Color("#40a02b"), // Green
		StatusFailure:           lipgloss.Color("#d20f39"), // Red
		StatusRunning:           lipgloss.Color("#df8e1d"), // Yellow
		StatusQueued:            lipgloss.Color("#9ca0b0"), // Overlay0
		StatusCancelled:         lipgloss.Color("#fe640b"), // Peach
		StatusBarBg:             lipgloss.Color("#e6e9ef"), // Mantle
		StatusBarFg:             lipgloss.Color("#4c4f69"), // Text
		ConfirmBorder:           lipgloss.Color("#fe640b"), // Peach
		HelpBorder:              lipgloss.Color("#04a5e5"), // Sky
		LogTimestamp:            lipgloss.Color("#04a5e5"), // Sky
		LogGroup:                lipgloss.Color("#40a02b"), // Green
		LogEndGroup:             lipgloss.Color("#179299"), // Teal
		LogError:                lipgloss.Color("#d20f39"), // Red
		LogWarning:              lipgloss.Color("#fe640b"), // Peach
		LogNotice:               lipgloss.Color("#1e66f5"), // Blue
		LogErrorKeyword:         lipgloss.Color("#e64553"), // Maroon
		LogWarningKeyword:       lipgloss.Color("#df8e1d"), // Yellow
		LogSuccessKeyword:       lipgloss.Color("#40a02b"), // Green
	}

	// DefaultTheme (Original classic high-contrast green/cyan theme)
	DefaultTheme = Theme{
		Name:                    "default",
		FocusedColor:            lipgloss.Color("#00FF00"),
		UnfocusedColor:          lipgloss.Color("#666666"),
		SelectedItemFg:          lipgloss.Color("#FFFFFF"),
		SelectedItemBg:          lipgloss.Color("#0066CC"),
		UnfocusedSelectedItemFg: lipgloss.Color("#CCCCCC"),
		UnfocusedSelectedItemBg: lipgloss.Color("#444444"),
		CursorFg:                lipgloss.Color("#00FF00"),
		NormalItemFg:            lipgloss.Color("#AAAAAA"),
		StatusSuccess:           lipgloss.Color("#00FF00"),
		StatusFailure:           lipgloss.Color("#FF0000"),
		StatusRunning:           lipgloss.Color("#FFFF00"),
		StatusQueued:            lipgloss.Color("#888888"),
		StatusCancelled:         lipgloss.Color("#FF8800"),
		StatusBarBg:             lipgloss.Color("#333333"),
		StatusBarFg:             lipgloss.Color("#FFFFFF"),
		ConfirmBorder:           lipgloss.Color("#FF8800"),
		HelpBorder:              lipgloss.Color("#00FFFF"),
		LogTimestamp:            lipgloss.Color("#00FFFF"),
		LogGroup:                lipgloss.Color("#00FF00"),
		LogEndGroup:             lipgloss.Color("#006600"),
		LogError:                lipgloss.Color("#FF0000"),
		LogWarning:              lipgloss.Color("#FF8800"),
		LogNotice:               lipgloss.Color("#00FFFF"),
		LogErrorKeyword:         lipgloss.Color("#FF6666"),
		LogWarningKeyword:       lipgloss.Color("#FFAA00"),
		LogSuccessKeyword:       lipgloss.Color("#66FF66"),
	}
)

// AvailableThemes returns a map of all built-in theme presets.
func AvailableThemes() map[string]Theme {
	return map[string]Theme{
		"catppuccin-mocha":     CatppuccinMocha,
		"mocha":                CatppuccinMocha,
		"catppuccin-macchiato": CatppuccinMacchiato,
		"macchiato":            CatppuccinMacchiato,
		"catppuccin-frappe":    CatppuccinFrappe,
		"frappe":               CatppuccinFrappe,
		"catppuccin-latte":     CatppuccinLatte,
		"latte":                CatppuccinLatte,
		"default":              DefaultTheme,
		"classic":              DefaultTheme,
	}
}

// GetTheme resolves a theme by name. Defaults to CatppuccinMocha if unknown.
func GetTheme(name string) Theme {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if theme, ok := AvailableThemes()[normalized]; ok {
		return theme
	}
	return CatppuccinMocha
}
