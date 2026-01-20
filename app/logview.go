package app

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// LogViewport wraps a viewport with autoscroll functionality.
// It automatically scrolls to the bottom when new content is added,
// unless the user has manually scrolled up.
type LogViewport struct {
	viewport   viewport.Model
	autoscroll bool
}

// NewLogViewport creates a new LogViewport with the specified dimensions.
func NewLogViewport(width, height int) *LogViewport {
	vp := viewport.New(width, height)
	return &LogViewport{
		viewport:   vp,
		autoscroll: true,
	}
}

// SetContent sets the content of the viewport.
// If autoscroll is enabled, it will scroll to the bottom.
func (lv *LogViewport) SetContent(content string) {
	lv.viewport.SetContent(content)
	if lv.autoscroll {
		lv.viewport.GotoBottom()
	}
}

// SetSize resizes the viewport.
func (lv *LogViewport) SetSize(width, height int) {
	lv.viewport.Width = width
	lv.viewport.Height = height
}

// View returns the rendered content of the viewport.
func (lv *LogViewport) View() string {
	return lv.viewport.View()
}

// Update handles messages and updates the viewport state.
// Returns the updated LogViewport and any commands.
func (lv *LogViewport) Update(msg tea.Msg) (*LogViewport, tea.Cmd) {
	if msg == nil {
		return lv, nil
	}

	var cmd tea.Cmd
	lv.viewport, cmd = lv.viewport.Update(msg)

	// Update autoscroll based on position
	lv.autoscroll = lv.isAtBottom()

	return lv, cmd
}

// isAtBottom returns true if the viewport is scrolled to the bottom.
func (lv *LogViewport) isAtBottom() bool {
	return lv.viewport.AtBottom()
}

// ScrollUp scrolls the viewport up by one line.
func (lv *LogViewport) ScrollUp() {
	lv.viewport.ScrollUp(1)
	lv.autoscroll = false
}

// ScrollDown scrolls the viewport down by one line.
func (lv *LogViewport) ScrollDown() {
	lv.viewport.ScrollDown(1)
	lv.autoscroll = lv.isAtBottom()
}

// GotoTop scrolls to the top of the content.
func (lv *LogViewport) GotoTop() {
	lv.viewport.GotoTop()
	lv.autoscroll = false
}
