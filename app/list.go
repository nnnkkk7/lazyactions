package app

import (
	"sync"
)

// FilteredList is a generic thread-safe list with filtering capability.
// It follows the lazydocker FilteredList[T] pattern for filterable lists
// used in TUI applications.
type FilteredList[T any] struct {
	mu            sync.RWMutex
	allItems      []T
	filtered      []T
	filter        string
	selectedIdx   int
	scrollOffset  int
	visibleHeight int
	matchFn       func(item T, filter string) bool
}

// NewFilteredList creates a new FilteredList with the provided match function.
// The matchFn is used to determine if an item matches the current filter.
// Panics if matchFn is nil.
func NewFilteredList[T any](matchFn func(T, string) bool) *FilteredList[T] {
	if matchFn == nil {
		panic("matchFn cannot be nil")
	}
	return &FilteredList[T]{
		allItems:    make([]T, 0),
		filtered:    make([]T, 0),
		matchFn:     matchFn,
		selectedIdx: 0,
	}
}

// SetItems sets the items in the list and applies the current filter.
// This replaces any existing items.
func (l *FilteredList[T]) SetItems(items []T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if items == nil {
		l.allItems = make([]T, 0)
	} else {
		l.allItems = items
	}
	l.applyFilter()
}

// SetFilter sets the filter string and refilters the items.
// An empty filter shows all items.
func (l *FilteredList[T]) SetFilter(filter string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.filter = filter
	l.applyFilter()
}

// applyFilter filters allItems based on the current filter.
// Must be called with the lock held.
func (l *FilteredList[T]) applyFilter() {
	if l.filter == "" {
		// No filter, show all items
		l.filtered = l.allItems
	} else {
		// Apply filter
		l.filtered = make([]T, 0)
		for _, item := range l.allItems {
			if l.matchFn(item, l.filter) {
				l.filtered = append(l.filtered, item)
			}
		}
	}

	// Clamp selectedIdx to valid range
	l.clampSelectedIndex()
	l.clampScrollOffset()
}

// clampSelectedIndex ensures selectedIdx is within valid bounds.
// Must be called with the lock held.
func (l *FilteredList[T]) clampSelectedIndex() {
	maxIdx := len(l.filtered) - 1
	if maxIdx < 0 {
		l.selectedIdx = 0
		return
	}
	if l.selectedIdx > maxIdx {
		l.selectedIdx = maxIdx
	} else if l.selectedIdx < 0 {
		l.selectedIdx = 0
	}
}

// clampScrollOffset adjusts scrollOffset so the selected item is visible.
// Must be called with the lock held.
func (l *FilteredList[T]) clampScrollOffset() {
	if l.visibleHeight <= 0 {
		return
	}
	// If selected item is above the visible window, scroll up
	if l.selectedIdx < l.scrollOffset {
		l.scrollOffset = l.selectedIdx
	}
	// If selected item is below the visible window, scroll down
	if l.selectedIdx >= l.scrollOffset+l.visibleHeight {
		l.scrollOffset = l.selectedIdx - l.visibleHeight + 1
	}
	// Ensure offset is not negative
	if l.scrollOffset < 0 {
		l.scrollOffset = 0
	}
	// Ensure we don't scroll past the end
	maxOffset := len(l.filtered) - l.visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if l.scrollOffset > maxOffset {
		l.scrollOffset = maxOffset
	}
}

// SetVisibleHeight sets the number of items visible in the panel.
// This is used to calculate the scroll offset.
func (l *FilteredList[T]) SetVisibleHeight(h int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.visibleHeight = h
	l.clampScrollOffset()
}

// ScrollOffset returns the current scroll offset (index of the first visible item).
func (l *FilteredList[T]) ScrollOffset() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.scrollOffset
}

// VisibleItems returns the slice of items currently visible in the panel.
func (l *FilteredList[T]) VisibleItems() []T {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.filtered) == 0 {
		return make([]T, 0)
	}
	start := l.scrollOffset
	if start >= len(l.filtered) {
		return make([]T, 0)
	}
	end := start + l.visibleHeight
	if end > len(l.filtered) || l.visibleHeight <= 0 {
		end = len(l.filtered)
	}
	return l.filtered[start:end]
}

// Items returns the filtered items.
// Returns an empty slice (not nil) if there are no items.
func (l *FilteredList[T]) Items() []T {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.filtered == nil {
		return make([]T, 0)
	}
	return l.filtered
}

// Selected returns the currently selected item and true, or zero value and false
// if the list is empty.
func (l *FilteredList[T]) Selected() (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.filtered) == 0 {
		var zero T
		return zero, false
	}
	return l.filtered[l.selectedIdx], true
}

// SelectNext moves the selection to the next item.
// Does nothing if already at the last item or if the list is empty.
func (l *FilteredList[T]) SelectNext() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.filtered) == 0 {
		return
	}
	if l.selectedIdx < len(l.filtered)-1 {
		l.selectedIdx++
	}
	l.clampScrollOffset()
}

// SelectPrev moves the selection to the previous item.
// Does nothing if already at the first item or if the list is empty.
func (l *FilteredList[T]) SelectPrev() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.filtered) == 0 {
		return
	}
	if l.selectedIdx > 0 {
		l.selectedIdx--
	}
	l.clampScrollOffset()
}

// SelectedIndex returns the index of the currently selected item.
func (l *FilteredList[T]) SelectedIndex() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.selectedIdx
}

// Len returns the number of filtered items.
func (l *FilteredList[T]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return len(l.filtered)
}

// Select sets the selection to the specified index.
// Does nothing if the index is out of bounds.
func (l *FilteredList[T]) Select(idx int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if idx >= 0 && idx < len(l.filtered) {
		l.selectedIdx = idx
	}
	l.clampScrollOffset()
}

// Reset clears the filter and resets the selection to the first item.
func (l *FilteredList[T]) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.filter = ""
	l.selectedIdx = 0
	l.scrollOffset = 0
	l.applyFilter()
}
