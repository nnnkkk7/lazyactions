package app

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Theme != "catppuccin-mocha" {
		t.Errorf("DefaultConfig().Theme = %q; want %q", cfg.Theme, "catppuccin-mocha")
	}
}

func TestConfigYamlParsing(t *testing.T) {
	yamlData := `
theme: catppuccin-latte
customTheme:
  focusedColor: "#123456"
  statusSuccess: "#654321"
`
	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	if err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	if cfg.Theme != "catppuccin-latte" {
		t.Errorf("cfg.Theme = %q; want %q", cfg.Theme, "catppuccin-latte")
	}

	if cfg.CustomTheme.FocusedColor != "#123456" {
		t.Errorf("cfg.CustomTheme.FocusedColor = %q; want %q", cfg.CustomTheme.FocusedColor, "#123456")
	}

	if cfg.CustomTheme.StatusSuccess != "#654321" {
		t.Errorf("cfg.CustomTheme.StatusSuccess = %q; want %q", cfg.CustomTheme.StatusSuccess, "#654321")
	}
}

func TestResolveTheme_CustomOverrides(t *testing.T) {
	cfg := &Config{
		Theme: "catppuccin-mocha",
		CustomTheme: CustomThemeConfig{
			FocusedColor: "#FF00FF",
		},
	}

	resolved := cfg.ResolveTheme()
	if resolved.Name != "catppuccin-mocha" {
		t.Errorf("resolved.Name = %q; want %q", resolved.Name, "catppuccin-mocha")
	}

	if string(resolved.FocusedColor) != "#FF00FF" {
		t.Errorf("resolved.FocusedColor = %q; want %q", string(resolved.FocusedColor), "#FF00FF")
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	// Set XDG_CONFIG_HOME to a temporary non-existent dir
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed for non-existent file: %v", err)
	}

	if cfg.Theme != "catppuccin-mocha" {
		t.Errorf("cfg.Theme = %q; want %q", cfg.Theme, "catppuccin-mocha")
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir := filepath.Join(tempDir, "lazyactions")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configContent := `
theme: catppuccin-frappe
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if cfg.Theme != "catppuccin-frappe" {
		t.Errorf("cfg.Theme = %q; want %q", cfg.Theme, "catppuccin-frappe")
	}
}
