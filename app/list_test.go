package app

import (
	"strings"
	"sync"
	"testing"
)

// =============================================================================
// Test Helper Types
// =============================================================================

// testItem is a simple struct for testing FilteredList
type testItem struct {
	Name string
	ID   int
}

// testMatchFn is a standard match function for testItem
func testMatchFn(item testItem, filter string) bool {
	return strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter))
}

// =============================================================================
// NewFilteredList Tests
// =============================================================================

func TestNewFilteredList_CreatesEmptyList(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	if list == nil {
		t.Fatal("NewFilteredList returned nil")
	}

	if list.Len() != 0 {
		t.Errorf("Len() = %d, want 0", list.Len())
	}

	items := list.Items()
	if len(items) != 0 {
		t.Errorf("Items() length = %d, want 0", len(items))
	}
}

func TestNewFilteredList_InitializesSelectedIndexToZero(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", list.SelectedIndex())
	}
}

func TestNewFilteredList_NilMatchFnPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewFilteredList(nil) should panic")
		}
	}()

	NewFilteredList[testItem](nil)
}

// =============================================================================
// SetItems Tests
// =============================================================================

func TestSetItems_SetsAllItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	items := []testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	}
	list.SetItems(items)

	got := list.Items()
	if len(got) != len(items) {
		t.Errorf("Items() length = %d, want %d", len(got), len(items))
	}

	for i, item := range got {
		if item.Name != items[i].Name || item.ID != items[i].ID {
			t.Errorf("Items()[%d] = %+v, want %+v", i, item, items[i])
		}
	}
}

func TestSetItems_ReplacesExistingItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	// Set initial items
	list.SetItems([]testItem{
		{Name: "Old1", ID: 1},
		{Name: "Old2", ID: 2},
	})

	// Replace with new items
	newItems := []testItem{
		{Name: "New1", ID: 10},
		{Name: "New2", ID: 20},
		{Name: "New3", ID: 30},
	}
	list.SetItems(newItems)

	got := list.Items()
	if len(got) != 3 {
		t.Errorf("Items() length = %d, want 3", len(got))
	}
	if got[0].Name != "New1" {
		t.Errorf("Items()[0].Name = %q, want %q", got[0].Name, "New1")
	}
}

func TestSetItems_EmptySliceClearsItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})

	list.SetItems([]testItem{})

	if list.Len() != 0 {
		t.Errorf("Len() = %d, want 0 after setting empty slice", list.Len())
	}
}

func TestSetItems_NilSliceClearsItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
	})

	list.SetItems(nil)

	if list.Len() != 0 {
		t.Errorf("Len() = %d, want 0 after setting nil slice", list.Len())
	}
}

func TestSetItems_AppliesExistingFilter(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	// Set filter first
	list.SetFilter("alpha")

	// Then set items
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "AlphaTwo", ID: 3},
	})

	got := list.Items()
	if len(got) != 2 {
		t.Errorf("Items() length = %d, want 2 (only alpha matches)", len(got))
	}
}

// =============================================================================
// SetFilter Tests
// =============================================================================

func TestSetFilter_FiltersItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
		{Name: "AlphaBeta", ID: 4},
	})

	list.SetFilter("alpha")

	got := list.Items()
	if len(got) != 2 {
		t.Errorf("Items() length = %d, want 2", len(got))
	}

	for _, item := range got {
		if !strings.Contains(strings.ToLower(item.Name), "alpha") {
			t.Errorf("Filtered item %q should contain 'alpha'", item.Name)
		}
	}
}

func TestSetFilter_EmptyFilterShowsAllItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	items := []testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	}
	list.SetItems(items)

	// Apply filter
	list.SetFilter("alpha")
	if list.Len() != 1 {
		t.Errorf("After filter, Len() = %d, want 1", list.Len())
	}

	// Clear filter
	list.SetFilter("")
	if list.Len() != 3 {
		t.Errorf("After clearing filter, Len() = %d, want 3", list.Len())
	}
}

func TestSetFilter_NoMatchesReturnsEmptyList(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})

	list.SetFilter("xyz")

	if list.Len() != 0 {
		t.Errorf("Len() = %d, want 0 for non-matching filter", list.Len())
	}
}

func TestSetFilter_CaseInsensitive(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "ALPHA", ID: 2},
		{Name: "alpha", ID: 3},
	})

	list.SetFilter("ALPHA")

	if list.Len() != 3 {
		t.Errorf("Len() = %d, want 3 (case-insensitive match)", list.Len())
	}
}

func TestSetFilter_ClampsSelectedIndex(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
		{Name: "Delta", ID: 4},
	})

	// Select last item
	list.SelectNext() // 1
	list.SelectNext() // 2
	list.SelectNext() // 3
	if list.SelectedIndex() != 3 {
		t.Errorf("SelectedIndex() = %d, want 3", list.SelectedIndex())
	}

	// Filter to only 2 items
	list.SetFilter("a") // Alpha, Gamma, Delta match

	// Index should be clamped to valid range
	idx := list.SelectedIndex()
	if idx < 0 || idx >= list.Len() {
		t.Errorf("SelectedIndex() = %d, out of valid range [0, %d)", idx, list.Len())
	}
}

func TestSetFilter_ClampsToZeroWhenEmpty(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})
	list.SelectNext() // Select index 1

	// Filter to nothing
	list.SetFilter("xyz")

	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0 when list is empty", list.SelectedIndex())
	}
}

// =============================================================================
// Items Tests
// =============================================================================

func TestItems_ReturnsFilteredItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})

	items := list.Items()
	if len(items) != 2 {
		t.Errorf("Items() length = %d, want 2", len(items))
	}
}

func TestItems_ReturnsEmptySliceWhenNoItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	items := list.Items()
	if items == nil {
		t.Error("Items() should return empty slice, not nil")
	}
	if len(items) != 0 {
		t.Errorf("Items() length = %d, want 0", len(items))
	}
}

// =============================================================================
// Selected Tests
// =============================================================================

func TestSelected_ReturnsSelectedItem(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	item, ok := list.Selected()
	if !ok {
		t.Error("Selected() returned false, want true")
	}
	if item.Name != "Alpha" {
		t.Errorf("Selected() = %+v, want Alpha", item)
	}
}

func TestSelected_ReturnsFalseWhenEmpty(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	_, ok := list.Selected()
	if ok {
		t.Error("Selected() returned true for empty list, want false")
	}
}

func TestSelected_ReturnsFalseWhenFilteredEmpty(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
	})

	list.SetFilter("xyz")

	_, ok := list.Selected()
	if ok {
		t.Error("Selected() returned true for empty filtered list, want false")
	}
}

func TestSelected_ReturnsCorrectItemAfterNavigation(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	list.SelectNext()

	item, ok := list.Selected()
	if !ok {
		t.Error("Selected() returned false, want true")
	}
	if item.Name != "Beta" {
		t.Errorf("Selected() = %+v, want Beta", item)
	}
}

// =============================================================================
// SelectNext Tests
// =============================================================================

func TestSelectNext_IncrementsIndex(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	if list.SelectedIndex() != 0 {
		t.Errorf("Initial SelectedIndex() = %d, want 0", list.SelectedIndex())
	}

	list.SelectNext()

	if list.SelectedIndex() != 1 {
		t.Errorf("After SelectNext(), SelectedIndex() = %d, want 1", list.SelectedIndex())
	}
}

func TestSelectNext_StopsAtLastItem(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})

	list.SelectNext() // 1
	list.SelectNext() // still 1 (at end)
	list.SelectNext() // still 1

	if list.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1 (should stop at last)", list.SelectedIndex())
	}
}

func TestSelectNext_NoOpWhenEmpty(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	// Should not panic on empty list
	list.SelectNext()

	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", list.SelectedIndex())
	}
}

// =============================================================================
// SelectPrev Tests
// =============================================================================

func TestSelectPrev_DecrementsIndex(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	list.SelectNext() // 1
	list.SelectNext() // 2

	list.SelectPrev()

	if list.SelectedIndex() != 1 {
		t.Errorf("After SelectPrev(), SelectedIndex() = %d, want 1", list.SelectedIndex())
	}
}

func TestSelectPrev_StopsAtFirstItem(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})

	list.SelectPrev() // still 0
	list.SelectPrev() // still 0

	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0 (should stop at first)", list.SelectedIndex())
	}
}

func TestSelectPrev_NoOpWhenEmpty(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	// Should not panic on empty list
	list.SelectPrev()

	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", list.SelectedIndex())
	}
}

// =============================================================================
// Select Tests
// =============================================================================

func TestSelect_SetsIndex(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	list.Select(2)

	if list.SelectedIndex() != 2 {
		t.Errorf("After Select(2), SelectedIndex() = %d, want 2", list.SelectedIndex())
	}
}

func TestSelect_IgnoresOutOfBoundsIndex(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})

	list.Select(1) // valid
	list.Select(5) // out of bounds - should be ignored

	if list.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1 (out of bounds should be ignored)", list.SelectedIndex())
	}
}

func TestSelect_IgnoresNegativeIndex(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})

	list.Select(1)  // valid
	list.Select(-1) // negative - should be ignored

	if list.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1 (negative index should be ignored)", list.SelectedIndex())
	}
}

// =============================================================================
// SelectedIndex Tests
// =============================================================================

func TestSelectedIndex_ReturnsCurrentIndex(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	tests := []struct {
		navigations int
		wantIdx     int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 2}, // clamped at max
	}

	for _, tt := range tests {
		list := NewFilteredList(testMatchFn)
		list.SetItems([]testItem{
			{Name: "Alpha", ID: 1},
			{Name: "Beta", ID: 2},
			{Name: "Gamma", ID: 3},
		})

		for i := 0; i < tt.navigations; i++ {
			list.SelectNext()
		}

		if list.SelectedIndex() != tt.wantIdx {
			t.Errorf("After %d SelectNext(), SelectedIndex() = %d, want %d",
				tt.navigations, list.SelectedIndex(), tt.wantIdx)
		}
	}
}

// =============================================================================
// Len Tests
// =============================================================================

func TestLen_ReturnsFilteredLength(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	if list.Len() != 3 {
		t.Errorf("Len() = %d, want 3", list.Len())
	}

	list.SetFilter("alph") // Only Alpha matches

	if list.Len() != 1 {
		t.Errorf("After filter, Len() = %d, want 1", list.Len())
	}
}

func TestLen_ReturnsZeroWhenEmpty(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	if list.Len() != 0 {
		t.Errorf("Len() = %d, want 0", list.Len())
	}
}

// =============================================================================
// Reset Tests
// =============================================================================

func TestReset_ClearsFilterAndResetsSelection(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	// Apply filter and navigate
	list.SetFilter("a")
	list.SelectNext()

	// Reset
	list.Reset()

	// Filter should be cleared
	if list.Len() != 3 {
		t.Errorf("After Reset(), Len() = %d, want 3 (all items)", list.Len())
	}

	// Selection should be reset
	if list.SelectedIndex() != 0 {
		t.Errorf("After Reset(), SelectedIndex() = %d, want 0", list.SelectedIndex())
	}
}

func TestReset_WorksOnEmptyList(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	// Should not panic
	list.Reset()

	if list.Len() != 0 {
		t.Errorf("Len() = %d, want 0", list.Len())
	}
	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", list.SelectedIndex())
	}
}

// =============================================================================
// Thread Safety Tests (Concurrent Access)
// =============================================================================

func TestFilteredList_ConcurrentSetItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			items := []testItem{
				{Name: "Item", ID: id},
			}
			list.SetItems(items)
		}(i)
	}

	wg.Wait()

	// Should complete without race conditions
	// The actual items don't matter, we're testing for data races
	_ = list.Items()
}

func TestFilteredList_ConcurrentSetFilter(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			filters := []string{"", "a", "b", "alpha"}
			list.SetFilter(filters[id%len(filters)])
		}(i)
	}

	wg.Wait()

	// Should complete without race conditions
	_ = list.Items()
}

func TestFilteredList_ConcurrentNavigation(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
		{Name: "Delta", ID: 4},
		{Name: "Epsilon", ID: 5},
	})

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				list.SelectNext()
			} else {
				list.SelectPrev()
			}
		}(i)
	}

	wg.Wait()

	// Index should be valid
	idx := list.SelectedIndex()
	if idx < 0 || idx >= list.Len() {
		t.Errorf("SelectedIndex() = %d, out of valid range [0, %d)", idx, list.Len())
	}
}

func TestFilteredList_ConcurrentMixedOperations(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			items := []testItem{
				{Name: "Alpha", ID: id},
				{Name: "Beta", ID: id + 1},
			}
			list.SetItems(items)
			list.SetFilter("a")
			list.SelectNext()
			list.Reset()
		}(i)
	}

	// Readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = list.Items()
			_ = list.Len()
			_ = list.SelectedIndex()
			_, _ = list.Selected()
		}()
	}

	wg.Wait()
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

func TestFilteredList_SingleItem(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Only", ID: 1},
	})

	if list.Len() != 1 {
		t.Errorf("Len() = %d, want 1", list.Len())
	}

	item, ok := list.Selected()
	if !ok {
		t.Error("Selected() returned false for single item list")
	}
	if item.Name != "Only" {
		t.Errorf("Selected().Name = %q, want %q", item.Name, "Only")
	}

	// Navigation should not change index
	list.SelectNext()
	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", list.SelectedIndex())
	}

	list.SelectPrev()
	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", list.SelectedIndex())
	}
}

func TestFilteredList_FilterThenSetItems(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	// Set filter before items
	list.SetFilter("b")

	// Then set items
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Bravo", ID: 3},
	})

	// Should have 2 items matching "b"
	if list.Len() != 2 {
		t.Errorf("Len() = %d, want 2", list.Len())
	}
}

func TestFilteredList_CustomMatchFunction(t *testing.T) {
	// Match items with even IDs when filter is set
	evenIDMatchFn := func(item testItem, filter string) bool {
		if filter == "" {
			return true
		}
		return item.ID%2 == 0
	}

	list := NewFilteredList(evenIDMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
		{Name: "Delta", ID: 4},
	})

	list.SetFilter("even")

	// Should match items with even IDs: Beta(2), Delta(4)
	if list.Len() != 2 {
		t.Errorf("Len() = %d, want 2", list.Len())
	}
}

func TestFilteredList_StringType(t *testing.T) {
	// Test with string type instead of struct
	stringMatchFn := func(item string, filter string) bool {
		return strings.Contains(strings.ToLower(item), strings.ToLower(filter))
	}

	list := NewFilteredList(stringMatchFn)

	list.SetItems([]string{"apple", "banana", "apricot", "cherry"})

	list.SetFilter("ap")

	items := list.Items()
	// "ap" matches: apple, apricot
	if len(items) != 2 {
		t.Errorf("Items() length = %d, want 2", len(items))
	}
}

func TestFilteredList_IntType(t *testing.T) {
	// Test with int type
	intMatchFn := func(item int, filter string) bool {
		if filter == "" {
			return true
		}
		return item%2 == 0 // filter even numbers when filter is set
	}

	list := NewFilteredList(intMatchFn)

	list.SetItems([]int{1, 2, 3, 4, 5, 6})

	if list.Len() != 6 {
		t.Errorf("Len() = %d, want 6", list.Len())
	}

	list.SetFilter("even")

	if list.Len() != 3 {
		t.Errorf("After filter, Len() = %d, want 3 (even numbers)", list.Len())
	}
}

// =============================================================================
// Selection Preservation Tests
// =============================================================================

func TestFilteredList_SelectionPreservedOnRefilter(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "AlphaBeta", ID: 3},
		{Name: "Gamma", ID: 4},
	})

	// Select index 2
	list.SelectNext() // 1
	list.SelectNext() // 2

	// Filter to reduce items but keep selected within range
	list.SetFilter("a") // Alpha, AlphaBeta, Gamma (index 2 still valid)

	idx := list.SelectedIndex()
	if idx >= list.Len() {
		t.Errorf("SelectedIndex() = %d, should be < %d", idx, list.Len())
	}
}

// =============================================================================
// SetVisibleHeight Tests
// =============================================================================

func TestSetVisibleHeight_SetsHeight(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})

	list.SetVisibleHeight(2)

	// visibleHeight is not directly accessible, but we can verify via VisibleItems
	items := list.VisibleItems()
	if len(items) != 2 {
		t.Errorf("VisibleItems() length = %d, want 2", len(items))
	}
}

func TestSetVisibleHeight_ClampsScrollOffset(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
		{Name: "Delta", ID: 4},
		{Name: "Epsilon", ID: 5},
	})

	list.SetVisibleHeight(2)

	// Navigate to last item so scrollOffset moves
	list.Select(4) // scrollOffset should be 3 (4 - 2 + 1)

	if list.ScrollOffset() != 3 {
		t.Errorf("ScrollOffset() = %d, want 3", list.ScrollOffset())
	}

	// Increase visible height — should clamp scroll offset down
	list.SetVisibleHeight(4) // maxOffset = 5 - 4 = 1

	got := list.ScrollOffset()
	if got > 1 {
		t.Errorf("After increasing visibleHeight, ScrollOffset() = %d, want <= 1", got)
	}
}

// =============================================================================
// ScrollOffset Tests
// =============================================================================

func TestScrollOffset_InitiallyZero(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})

	if list.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset() = %d, want 0", list.ScrollOffset())
	}
}

func TestScrollOffset_StaysZeroWhenNoScrollNeeded(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})
	list.SetVisibleHeight(5) // More visible space than items

	list.SelectNext() // Select index 1

	if list.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset() = %d, want 0 (enough room for all items)", list.ScrollOffset())
	}
}

// =============================================================================
// VisibleItems Tests
// =============================================================================

func TestVisibleItems_ReturnsCorrectWindow(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
		{Name: "Delta", ID: 4},
		{Name: "Epsilon", ID: 5},
	})
	list.SetVisibleHeight(3)

	items := list.VisibleItems()
	if len(items) != 3 {
		t.Fatalf("VisibleItems() length = %d, want 3", len(items))
	}
	if items[0].Name != "Alpha" {
		t.Errorf("VisibleItems()[0].Name = %q, want %q", items[0].Name, "Alpha")
	}
	if items[2].Name != "Gamma" {
		t.Errorf("VisibleItems()[2].Name = %q, want %q", items[2].Name, "Gamma")
	}
}

func TestVisibleItems_ShiftsWhenScrolled(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
		{Name: "Delta", ID: 4},
		{Name: "Epsilon", ID: 5},
	})
	list.SetVisibleHeight(3)

	// Navigate to item 3 (index 3) — should scroll so items 1,2,3 are visible
	list.Select(3)

	items := list.VisibleItems()
	if len(items) != 3 {
		t.Fatalf("VisibleItems() length = %d, want 3", len(items))
	}
	if items[0].Name != "Beta" {
		t.Errorf("VisibleItems()[0].Name = %q, want %q", items[0].Name, "Beta")
	}
	if items[2].Name != "Delta" {
		t.Errorf("VisibleItems()[2].Name = %q, want %q", items[2].Name, "Delta")
	}
}

func TestVisibleItems_ReturnsEmptyForEmptyList(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetVisibleHeight(5)

	items := list.VisibleItems()
	if items == nil {
		t.Error("VisibleItems() should return empty slice, not nil")
	}
	if len(items) != 0 {
		t.Errorf("VisibleItems() length = %d, want 0", len(items))
	}
}

func TestVisibleItems_ReturnsAllItemsWhenFewerThanHeight(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})
	list.SetVisibleHeight(10)

	items := list.VisibleItems()
	if len(items) != 2 {
		t.Errorf("VisibleItems() length = %d, want 2", len(items))
	}
}

func TestVisibleItems_ReturnsAllItemsWhenHeightZero(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
	})
	list.SetVisibleHeight(0)

	// visibleHeight 0 means no viewport constraint — returns all items
	items := list.VisibleItems()
	if len(items) != 3 {
		t.Errorf("VisibleItems() with height 0: length = %d, want 3", len(items))
	}
}

func TestVisibleItems_ReturnsAllItemsWhenHeightNotSet(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
	})

	// Never called SetVisibleHeight — default is 0
	items := list.VisibleItems()
	if len(items) != 2 {
		t.Errorf("VisibleItems() with no height set: length = %d, want 2", len(items))
	}
}

// =============================================================================
// SelectNext Scroll Offset Tests
// =============================================================================

func TestSelectNext_ScrollsDownWhenPastVisibleWindow(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "A", ID: 1},
		{Name: "B", ID: 2},
		{Name: "C", ID: 3},
		{Name: "D", ID: 4},
		{Name: "E", ID: 5},
	})
	list.SetVisibleHeight(3)

	// Navigate past visible window: items 0,1,2 visible initially
	list.SelectNext() // idx 1, offset 0
	list.SelectNext() // idx 2, offset 0
	list.SelectNext() // idx 3, offset should adjust to 1

	if list.ScrollOffset() != 1 {
		t.Errorf("ScrollOffset() = %d, want 1", list.ScrollOffset())
	}

	// Selected item should be visible in the window
	items := list.VisibleItems()
	found := false
	for _, item := range items {
		if item.Name == "D" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Selected item 'D' should be in VisibleItems()")
	}
}

func TestSelectNext_ScrollsToEndCorrectly(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "A", ID: 1},
		{Name: "B", ID: 2},
		{Name: "C", ID: 3},
		{Name: "D", ID: 4},
		{Name: "E", ID: 5},
	})
	list.SetVisibleHeight(3)

	// Navigate to last item
	for i := 0; i < 4; i++ {
		list.SelectNext()
	}

	if list.SelectedIndex() != 4 {
		t.Errorf("SelectedIndex() = %d, want 4", list.SelectedIndex())
	}

	// scrollOffset = 4 - 3 + 1 = 2
	if list.ScrollOffset() != 2 {
		t.Errorf("ScrollOffset() = %d, want 2", list.ScrollOffset())
	}
}

// =============================================================================
// SelectPrev Scroll Offset Tests
// =============================================================================

func TestSelectPrev_ScrollsUpWhenAboveVisibleWindow(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "A", ID: 1},
		{Name: "B", ID: 2},
		{Name: "C", ID: 3},
		{Name: "D", ID: 4},
		{Name: "E", ID: 5},
	})
	list.SetVisibleHeight(3)

	// Navigate to end first
	list.Select(4) // scrollOffset = 2, visible: C,D,E

	// Navigate back up
	list.SelectPrev() // idx 3, offset 2 (still in window)
	list.SelectPrev() // idx 2, offset 2 (at top of window)
	list.SelectPrev() // idx 1, offset should adjust to 1

	if list.ScrollOffset() != 1 {
		t.Errorf("ScrollOffset() = %d, want 1", list.ScrollOffset())
	}
}

func TestSelectPrev_ScrollsToTopCorrectly(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "A", ID: 1},
		{Name: "B", ID: 2},
		{Name: "C", ID: 3},
		{Name: "D", ID: 4},
		{Name: "E", ID: 5},
	})
	list.SetVisibleHeight(3)

	// Navigate to end
	list.Select(4)

	// Navigate all the way back to top
	for i := 0; i < 4; i++ {
		list.SelectPrev()
	}

	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", list.SelectedIndex())
	}
	if list.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset() = %d, want 0", list.ScrollOffset())
	}
}

// =============================================================================
// Select Scroll Offset Tests
// =============================================================================

func TestSelect_AdjustsScrollOffsetForArbitraryIndex(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "A", ID: 1},
		{Name: "B", ID: 2},
		{Name: "C", ID: 3},
		{Name: "D", ID: 4},
		{Name: "E", ID: 5},
		{Name: "F", ID: 6},
		{Name: "G", ID: 7},
		{Name: "H", ID: 8},
		{Name: "I", ID: 9},
		{Name: "J", ID: 10},
	})
	list.SetVisibleHeight(3)

	// Jump to middle
	list.Select(5) // offset = 5 - 3 + 1 = 3

	if list.ScrollOffset() != 3 {
		t.Errorf("ScrollOffset() = %d, want 3", list.ScrollOffset())
	}

	// Jump back to start
	list.Select(0) // offset = 0

	if list.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset() = %d, want 0", list.ScrollOffset())
	}

	// Jump to end
	list.Select(9) // offset = 9 - 3 + 1 = 7

	if list.ScrollOffset() != 7 {
		t.Errorf("ScrollOffset() = %d, want 7", list.ScrollOffset())
	}
}

// =============================================================================
// Reset Scroll Offset Tests
// =============================================================================

func TestReset_ResetsScrollOffset(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "A", ID: 1},
		{Name: "B", ID: 2},
		{Name: "C", ID: 3},
		{Name: "D", ID: 4},
		{Name: "E", ID: 5},
	})
	list.SetVisibleHeight(2)

	// Scroll to the end
	list.Select(4)

	if list.ScrollOffset() == 0 {
		t.Fatal("ScrollOffset() should be non-zero before Reset()")
	}

	list.Reset()

	if list.ScrollOffset() != 0 {
		t.Errorf("After Reset(), ScrollOffset() = %d, want 0", list.ScrollOffset())
	}
	if list.SelectedIndex() != 0 {
		t.Errorf("After Reset(), SelectedIndex() = %d, want 0", list.SelectedIndex())
	}
}

// =============================================================================
// SetFilter / SetItems Scroll Offset Clamping Tests
// =============================================================================

func TestSetFilter_ClampsScrollOffset(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Alpha", ID: 1},
		{Name: "Beta", ID: 2},
		{Name: "Gamma", ID: 3},
		{Name: "Delta", ID: 4},
		{Name: "Epsilon", ID: 5},
	})
	list.SetVisibleHeight(2)

	// Navigate to end
	list.Select(4) // scrollOffset = 3

	if list.ScrollOffset() != 3 {
		t.Fatalf("ScrollOffset() = %d, want 3 before filter", list.ScrollOffset())
	}

	// Filter to just 2 items — scrollOffset must be clamped
	list.SetFilter("a") // Alpha, Gamma, Delta match "a" = 3 items

	offset := list.ScrollOffset()
	maxOffset := list.Len() - 2 // Len - visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		t.Errorf("After filter, ScrollOffset() = %d, exceeds maxOffset %d", offset, maxOffset)
	}
}

func TestSetItems_ClampsScrollOffset(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "A", ID: 1},
		{Name: "B", ID: 2},
		{Name: "C", ID: 3},
		{Name: "D", ID: 4},
		{Name: "E", ID: 5},
	})
	list.SetVisibleHeight(2)

	// Navigate to end
	list.Select(4) // scrollOffset = 3

	// Replace with fewer items
	list.SetItems([]testItem{
		{Name: "X", ID: 1},
		{Name: "Y", ID: 2},
	})

	offset := list.ScrollOffset()
	if offset != 0 {
		t.Errorf("After replacing items, ScrollOffset() = %d, want 0", offset)
	}
}

// =============================================================================
// Scroll Edge Cases Tests
// =============================================================================

func TestScroll_SingleItem(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "Only", ID: 1},
	})
	list.SetVisibleHeight(3)

	if list.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset() = %d, want 0 for single item", list.ScrollOffset())
	}

	items := list.VisibleItems()
	if len(items) != 1 {
		t.Errorf("VisibleItems() length = %d, want 1", len(items))
	}
}

func TestScroll_EmptyListWithVisibleHeight(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetVisibleHeight(5)

	if list.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset() = %d, want 0 for empty list", list.ScrollOffset())
	}

	items := list.VisibleItems()
	if len(items) != 0 {
		t.Errorf("VisibleItems() length = %d, want 0 for empty list", len(items))
	}
}

func TestScroll_VisibleHeightExactlyMatchesItemCount(t *testing.T) {
	list := NewFilteredList(testMatchFn)
	list.SetItems([]testItem{
		{Name: "A", ID: 1},
		{Name: "B", ID: 2},
		{Name: "C", ID: 3},
	})
	list.SetVisibleHeight(3)

	// Navigate to last item — should not scroll since all fit
	list.Select(2)

	if list.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset() = %d, want 0 when all items fit", list.ScrollOffset())
	}

	items := list.VisibleItems()
	if len(items) != 3 {
		t.Errorf("VisibleItems() length = %d, want 3", len(items))
	}
}

// =============================================================================
// Concurrent Scroll Tests
// =============================================================================

func TestFilteredList_ConcurrentScrollOperations(t *testing.T) {
	list := NewFilteredList(testMatchFn)

	items := make([]testItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = testItem{Name: "Item", ID: i}
	}
	list.SetItems(items)
	list.SetVisibleHeight(10)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Writers: navigate and change visible height
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%3 == 0 {
				list.SelectNext()
			} else if id%3 == 1 {
				list.SelectPrev()
			} else {
				list.SetVisibleHeight(id%20 + 1)
			}
		}(i)
	}

	// Readers: read scroll state
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = list.ScrollOffset()
			_ = list.VisibleItems()
		}()
	}

	wg.Wait()

	// Verify scroll offset is valid
	offset := list.ScrollOffset()
	if offset < 0 {
		t.Errorf("ScrollOffset() = %d, should be >= 0", offset)
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkFilteredList_SetItems(b *testing.B) {
	list := NewFilteredList(testMatchFn)
	items := make([]testItem, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = testItem{Name: "Item", ID: i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.SetItems(items)
	}
}

func BenchmarkFilteredList_SetFilter(b *testing.B) {
	list := NewFilteredList(testMatchFn)
	items := make([]testItem, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = testItem{Name: "Item", ID: i}
	}
	list.SetItems(items)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.SetFilter("item")
	}
}

func BenchmarkFilteredList_ConcurrentAccess(b *testing.B) {
	list := NewFilteredList(testMatchFn)
	items := make([]testItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = testItem{Name: "Item", ID: i}
	}
	list.SetItems(items)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			list.Items()
			list.Selected()
			list.Len()
		}
	})
}
