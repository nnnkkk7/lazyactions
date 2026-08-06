package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Theme       string            `yaml:"theme"`       // Theme preset name (e.g. "catppuccin-mocha", "latte", "default")
	CustomTheme CustomThemeConfig `yaml:"customTheme"` // Optional fine-grained color overrides
}

// CustomThemeConfig allows overriding individual theme color tokens via YAML.
type CustomThemeConfig struct {
	FocusedColor            string `yaml:"focusedColor"`
	UnfocusedColor          string `yaml:"unfocusedColor"`
	SelectedItemFg         string `yaml:"selectedItemFg"`
	SelectedItemBg         string `yaml:"selectedItemBg"`
	UnfocusedSelectedItemFg string `yaml:"unfocusedSelectedItemFg"`
	UnfocusedSelectedItemBg string `yaml:"unfocusedSelectedItemBg"`
	CursorFg               string `yaml:"cursorFg"`
	NormalItemFg           string `yaml:"normalItemFg"`
	StatusSuccess          string `yaml:"statusSuccess"`
	StatusFailure          string `yaml:"statusFailure"`
	StatusRunning          string `yaml:"statusRunning"`
	StatusQueued           string `yaml:"statusQueued"`
	StatusCancelled        string `yaml:"statusCancelled"`
	StatusBarBg            string `yaml:"statusBarBg"`
	StatusBarFg            string `yaml:"statusBarFg"`
	ConfirmBorder          string `yaml:"confirmBorder"`
	HelpBorder             string `yaml:"helpBorder"`
	LogTimestamp           string `yaml:"logTimestamp"`
	LogGroup               string `yaml:"logGroup"`
	LogEndGroup            string `yaml:"logEndGroup"`
	LogError               string `yaml:"logError"`
	LogWarning             string `yaml:"logWarning"`
	LogNotice              string `yaml:"logNotice"`
	LogErrorKeyword        string `yaml:"logErrorKeyword"`
	LogWarningKeyword      string `yaml:"logWarningKeyword"`
	LogSuccessKeyword      string `yaml:"logSuccessKeyword"`
}

// DefaultConfig returns the default application configuration.
func DefaultConfig() *Config {
	return &Config{
		Theme: "catppuccin-mocha",
	}
}

// ConfigFilePath returns the path to the user's configuration file ($XDG_CONFIG_HOME/lazyactions/config.yml or ~/.config/lazyactions/config.yml).
func ConfigFilePath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "lazyactions", "config.yml")
}

// LoadConfig loads the user configuration file if it exists, or returns the default config.
func LoadConfig() (*Config, error) {
	path := ConfigFilePath()
	if path == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file at %s: %w", path, err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file at %s: %w", path, err)
	}

	return cfg, nil
}

// ResolveTheme converts Config into a resolved Theme struct, applying preset base and custom overrides.
func (c *Config) ResolveTheme() Theme {
	themeName := c.Theme
	if themeName == "" {
		themeName = "catppuccin-mocha"
	}
	theme := GetTheme(themeName)

	// Apply custom theme overrides if specified
	applyOverride := func(target *lipgloss.Color, hex string) {
		if hex != "" {
			*target = lipgloss.Color(hex)
		}
	}

	ct := c.CustomTheme
	applyOverride(&theme.FocusedColor, ct.FocusedColor)
	applyOverride(&theme.UnfocusedColor, ct.UnfocusedColor)
	applyOverride(&theme.SelectedItemFg, ct.SelectedItemFg)
	applyOverride(&theme.SelectedItemBg, ct.SelectedItemBg)
	applyOverride(&theme.UnfocusedSelectedItemFg, ct.UnfocusedSelectedItemFg)
	applyOverride(&theme.UnfocusedSelectedItemBg, ct.UnfocusedSelectedItemBg)
	applyOverride(&theme.CursorFg, ct.CursorFg)
	applyOverride(&theme.NormalItemFg, ct.NormalItemFg)
	applyOverride(&theme.StatusSuccess, ct.StatusSuccess)
	applyOverride(&theme.StatusFailure, ct.StatusFailure)
	applyOverride(&theme.StatusRunning, ct.StatusRunning)
	applyOverride(&theme.StatusQueued, ct.StatusQueued)
	applyOverride(&theme.StatusCancelled, ct.StatusCancelled)
	applyOverride(&theme.StatusBarBg, ct.StatusBarBg)
	applyOverride(&theme.StatusBarFg, ct.StatusBarFg)
	applyOverride(&theme.ConfirmBorder, ct.ConfirmBorder)
	applyOverride(&theme.HelpBorder, ct.HelpBorder)
	applyOverride(&theme.LogTimestamp, ct.LogTimestamp)
	applyOverride(&theme.LogGroup, ct.LogGroup)
	applyOverride(&theme.LogEndGroup, ct.LogEndGroup)
	applyOverride(&theme.LogError, ct.LogError)
	applyOverride(&theme.LogWarning, ct.LogWarning)
	applyOverride(&theme.LogNotice, ct.LogNotice)
	applyOverride(&theme.LogErrorKeyword, ct.LogErrorKeyword)
	applyOverride(&theme.LogWarningKeyword, ct.LogWarningKeyword)
	applyOverride(&theme.LogSuccessKeyword, ct.LogSuccessKeyword)

	return theme
}
