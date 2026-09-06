package app

import (
	"testing"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"catppuccin-mocha", "catppuccin-mocha"},
		{"mocha", "catppuccin-mocha"},
		{"catppuccin-macchiato", "catppuccin-macchiato"},
		{"macchiato", "catppuccin-macchiato"},
		{"catppuccin-frappe", "catppuccin-frappe"},
		{"frappe", "catppuccin-frappe"},
		{"catppuccin-latte", "catppuccin-latte"},
		{"latte", "catppuccin-latte"},
		{"default", "default"},
		{"classic", "default"},
		{"unknown-theme", "catppuccin-mocha"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			theme := GetTheme(tt.input)
			if theme.Name != tt.expected {
				t.Errorf("GetTheme(%q).Name = %q; want %q", tt.input, theme.Name, tt.expected)
			}
		})
	}
}

func TestAvailableThemes(t *testing.T) {
	themes := AvailableThemes()
	if len(themes) == 0 {
		t.Fatal("AvailableThemes() returned empty map")
	}

	expectedKeys := []string{"catppuccin-mocha", "catppuccin-macchiato", "catppuccin-frappe", "catppuccin-latte", "default"}
	for _, key := range expectedKeys {
		if _, ok := themes[key]; !ok {
			t.Errorf("AvailableThemes() missing expected key %q", key)
		}
	}
}

func TestApplyTheme(t *testing.T) {
	ApplyTheme(CatppuccinMocha)
	if CurrentTheme.Name != "catppuccin-mocha" {
		t.Errorf("CurrentTheme.Name = %q; want %q", CurrentTheme.Name, "catppuccin-mocha")
	}

	ApplyTheme(DefaultTheme)
	if CurrentTheme.Name != "default" {
		t.Errorf("CurrentTheme.Name = %q; want %q", CurrentTheme.Name, "default")
	}
}
