package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

// =============================================================================
// DefaultKeyMap Tests
// =============================================================================

func TestDefaultKeyMap_ReturnsNonEmptyKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Verify that the struct has all required bindings
	if len(km.Up.Keys()) == 0 {
		t.Error("Up binding has no keys")
	}
	if len(km.Down.Keys()) == 0 {
		t.Error("Down binding has no keys")
	}
	if len(km.Left.Keys()) == 0 {
		t.Error("Left binding has no keys")
	}
	if len(km.Right.Keys()) == 0 {
		t.Error("Right binding has no keys")
	}
}

func TestDefaultKeyMap_NavigationKeys(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name     string
		binding  key.Binding
		wantKeys []string
	}{
		{"Up", km.Up, []string{"k", "up"}},
		{"Down", km.Down, []string{"j", "down"}},
		{"Left", km.Left, []string{"h", "left"}},
		{"Right", km.Right, []string{"l", "right"}},
		{"Tab", km.Tab, []string{"tab"}},
		{"ShiftTab", km.ShiftTab, []string{"shift+tab"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()
			if len(keys) != len(tt.wantKeys) {
				t.Errorf("%s binding has %d keys, want %d keys",
					tt.name, len(keys), len(tt.wantKeys))
				return
			}
			for i, wantKey := range tt.wantKeys {
				if keys[i] != wantKey {
					t.Errorf("%s binding keys[%d] = %q, want %q",
						tt.name, i, keys[i], wantKey)
				}
			}
		})
	}
}

func TestDefaultKeyMap_ActionKeys(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name     string
		binding  key.Binding
		wantKeys []string
	}{
		{"Enter", km.Enter, []string{"enter"}},
		{"Trigger", km.Trigger, []string{"t"}},
		{"Cancel", km.Cancel, []string{"c"}},
		{"Rerun", km.Rerun, []string{"r"}},
		{"RerunFailed", km.RerunFailed, []string{"R"}},
		{"Yank", km.Yank, []string{"y"}},
		{"Filter", km.Filter, []string{"/"}},
		{"Refresh", km.Refresh, []string{"ctrl+r"}},
		{"FullLog", km.FullLog, []string{"L"}},
		{"Help", km.Help, []string{"?"}},
		{"Quit", km.Quit, []string{"q"}},
		{"Escape", km.Escape, []string{"esc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()
			if len(keys) != len(tt.wantKeys) {
				t.Errorf("%s binding has %d keys, want %d keys",
					tt.name, len(keys), len(tt.wantKeys))
				return
			}
			for i, wantKey := range tt.wantKeys {
				if keys[i] != wantKey {
					t.Errorf("%s binding keys[%d] = %q, want %q",
						tt.name, i, keys[i], wantKey)
				}
			}
		})
	}
}

func TestDefaultKeyMap_HelpText(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name     string
		binding  key.Binding
		wantKey  string
		wantDesc string
	}{
		{"Up", km.Up, "k/↑", "move up"},
		{"Down", km.Down, "j/↓", "move down"},
		{"Left", km.Left, "h/←", "previous pane"},
		{"Right", km.Right, "l/→", "next pane"},
		{"Tab", km.Tab, "tab", "next pane"},
		{"ShiftTab", km.ShiftTab, "shift+tab", "previous pane"},
		{"Enter", km.Enter, "enter", "select/confirm"},
		{"Trigger", km.Trigger, "t", "trigger workflow"},
		{"Cancel", km.Cancel, "c", "cancel run"},
		{"Rerun", km.Rerun, "r", "rerun workflow"},
		{"RerunFailed", km.RerunFailed, "R", "rerun failed jobs"},
		{"Yank", km.Yank, "y", "copy to clipboard"},
		{"Filter", km.Filter, "/", "filter"},
		{"Refresh", km.Refresh, "ctrl+r", "refresh"},
		{"FullLog", km.FullLog, "L", "full log view"},
		{"Help", km.Help, "?", "help"},
		{"Quit", km.Quit, "q", "quit"},
		{"Escape", km.Escape, "esc", "back/cancel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.binding.Help()
			if help.Key != tt.wantKey {
				t.Errorf("%s help key = %q, want %q", tt.name, help.Key, tt.wantKey)
			}
			if help.Desc != tt.wantDesc {
				t.Errorf("%s help desc = %q, want %q", tt.name, help.Desc, tt.wantDesc)
			}
		})
	}
}

func TestDefaultKeyMap_AllBindingsEnabled(t *testing.T) {
	km := DefaultKeyMap()

	bindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Up", km.Up},
		{"Down", km.Down},
		{"Left", km.Left},
		{"Right", km.Right},
		{"Tab", km.Tab},
		{"ShiftTab", km.ShiftTab},
		{"Enter", km.Enter},
		{"Trigger", km.Trigger},
		{"Cancel", km.Cancel},
		{"Rerun", km.Rerun},
		{"RerunFailed", km.RerunFailed},
		{"Yank", km.Yank},
		{"Filter", km.Filter},
		{"Refresh", km.Refresh},
		{"FullLog", km.FullLog},
		{"Help", km.Help},
		{"Quit", km.Quit},
		{"Escape", km.Escape},
	}

	for _, b := range bindings {
		t.Run(b.name, func(t *testing.T) {
			if !b.binding.Enabled() {
				t.Errorf("%s binding is disabled, want enabled", b.name)
			}
		})
	}
}

func TestDefaultKeyMap_CaseSensitiveKeys(t *testing.T) {
	km := DefaultKeyMap()

	// Rerun should be lowercase 'r'
	rerunKeys := km.Rerun.Keys()
	for _, k := range rerunKeys {
		if k == "R" {
			t.Errorf("Rerun binding should use lowercase 'r', not 'R'")
		}
	}

	// RerunFailed should be uppercase 'R'
	rerunFailedKeys := km.RerunFailed.Keys()
	found := false
	for _, k := range rerunFailedKeys {
		if k == "R" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("RerunFailed binding should use uppercase 'R'")
	}

	// FullLog should be uppercase 'L'
	fullLogKeys := km.FullLog.Keys()
	found = false
	for _, k := range fullLogKeys {
		if k == "L" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("FullLog binding should use uppercase 'L'")
	}
}

func TestDefaultKeyMap_KeyMapMatches(t *testing.T) {
	km := DefaultKeyMap()

	// Test that key.Matches works correctly for each binding
	tests := []struct {
		name    string
		binding key.Binding
		testKey string
	}{
		{"Up_k", km.Up, "k"},
		{"Up_up", km.Up, "up"},
		{"Down_j", km.Down, "j"},
		{"Down_down", km.Down, "down"},
		{"Left_h", km.Left, "h"},
		{"Left_left", km.Left, "left"},
		{"Right_l", km.Right, "l"},
		{"Right_right", km.Right, "right"},
		{"Quit_q", km.Quit, "q"},
		{"Help_question", km.Help, "?"},
		{"Filter_slash", km.Filter, "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !keyContains(tt.binding.Keys(), tt.testKey) {
				t.Errorf("%s binding should contain key %q", tt.name, tt.testKey)
			}
		})
	}
}

// Helper function to check if a key is in the slice
func keyContains(keys []string, target string) bool {
	for _, k := range keys {
		if k == target {
			return true
		}
	}
	return false
}

func TestDefaultKeyMap_ConsistentReturns(t *testing.T) {
	// Ensure DefaultKeyMap returns consistent results
	km1 := DefaultKeyMap()
	km2 := DefaultKeyMap()

	// Check that both keymaps have the same keys
	if km1.Up.Keys()[0] != km2.Up.Keys()[0] {
		t.Errorf("DefaultKeyMap() should return consistent results")
	}
	if km1.Quit.Keys()[0] != km2.Quit.Keys()[0] {
		t.Errorf("DefaultKeyMap() should return consistent results")
	}
}

func TestKeyMapStruct_AllFieldsDefined(t *testing.T) {
	// This test ensures that the KeyMap struct has all expected fields
	km := DefaultKeyMap()

	// Access all fields to ensure they exist
	_ = km.Up
	_ = km.Down
	_ = km.Left
	_ = km.Right
	_ = km.Tab
	_ = km.ShiftTab
	_ = km.Enter
	_ = km.Trigger
	_ = km.Cancel
	_ = km.Rerun
	_ = km.RerunFailed
	_ = km.Yank
	_ = km.Filter
	_ = km.Refresh
	_ = km.FullLog
	_ = km.Help
	_ = km.Quit
	_ = km.Escape

	// If we get here without a compile error, all fields exist
}
