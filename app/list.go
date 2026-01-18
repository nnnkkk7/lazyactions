package app

import (
	"sync"
)

// FilteredList is a generic thread-safe list with filtering capability.
// It follows the lazydocker FilteredList[T] pattern for filterable lists
// used in TUI applications.
type FilteredList[T any] struct {
	mu          sync.RWMutex
	allItems    []T
	filtered    []T
	filter      string
	selectedIdx int
	matchFn     func(item T, filter string) bool
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

// Reset clears the filter and resets the selection to the first item.
func (l *FilteredList[T]) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.filter = ""
	l.selectedIdx = 0
	l.applyFilter()
}
