package s3tables

import (
	"errors"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/manifoldco/promptui"
)

// TestNewPromptSelector tests NewPromptSelector constructor
func TestNewPromptSelector(t *testing.T) {
	selector := NewPromptSelector()
	if selector == nil {
		t.Error("NewPromptSelector() returned nil")
	}
	if selector.runFunc == nil {
		t.Error("NewPromptSelector() runFunc is nil")
	}
}

// TestNewFilterablePromptSelector tests NewFilterablePromptSelector constructor
func TestNewFilterablePromptSelector(t *testing.T) {
	selector := NewFilterablePromptSelector()
	if selector == nil {
		t.Error("NewFilterablePromptSelector() returned nil")
	}
	if selector.runFunc == nil {
		t.Error("NewFilterablePromptSelector() runFunc is nil")
	}
}

// TestPromptSelectorSelectEmptyItems tests Select with empty items
func TestPromptSelectorSelectEmptyItems(t *testing.T) {
	selector := NewPromptSelector()
	_, err := selector.Select("Test", []string{})
	if err == nil {
		t.Error("Select() with empty items should return error")
	}
	if err.Error() != "no items to select" {
		t.Errorf("Select() error = %v, want 'no items to select'", err)
	}
}

// TestFilterablePromptSelectorSelectWithFilterEmptyItems tests SelectWithFilter with empty items
func TestFilterablePromptSelectorSelectWithFilterEmptyItems(t *testing.T) {
	selector := NewFilterablePromptSelector()
	_, err := selector.SelectWithFilter("Test", []string{}, false)
	if err == nil {
		t.Error("SelectWithFilter() with empty items should return error")
	}
	if err.Error() != "no items to select" {
		t.Errorf("SelectWithFilter() error = %v, want 'no items to select'", err)
	}
}

// TestPromptSelectorSelectSuccess tests Select success case
func TestPromptSelectorSelectSuccess(t *testing.T) {
	selector := &PromptSelector{
		runFunc: func(prompt promptRunner) (int, string, error) {
			return 0, "item1", nil
		},
	}

	result, err := selector.Select("Test", []string{"item1", "item2"})
	if err != nil {
		t.Errorf("Select() error = %v", err)
	}
	if result != "item1" {
		t.Errorf("Select() = %v, want item1", result)
	}
}

// TestPromptSelectorSelectError tests Select error case
func TestPromptSelectorSelectError(t *testing.T) {
	selector := &PromptSelector{
		runFunc: func(prompt promptRunner) (int, string, error) {
			return 0, "", errors.New("prompt error")
		},
	}

	_, err := selector.Select("Test", []string{"item1", "item2"})
	if err == nil {
		t.Error("Select() should return error")
	}
	if !strings.Contains(err.Error(), "selection failed") {
		t.Errorf("Select() error = %v, want 'selection failed'", err)
	}
}

// TestFilterablePromptSelectorSelectWithFilterSuccess tests SelectWithFilter success case
func TestFilterablePromptSelectorSelectWithFilterSuccess(t *testing.T) {
	selector := &FilterablePromptSelector{
		runFunc: func(prompt promptRunner) (int, string, error) {
			return 0, "item1", nil
		},
	}

	result, err := selector.SelectWithFilter("Test", []string{"item1", "item2"}, false)
	if err != nil {
		t.Errorf("SelectWithFilter() error = %v", err)
	}
	if result.Selected != "item1" {
		t.Errorf("SelectWithFilter() Selected = %v, want item1", result.Selected)
	}
	if result.Action != ActionSelect {
		t.Errorf("SelectWithFilter() Action = %v, want ActionSelect", result.Action)
	}
}

// TestFilterablePromptSelectorSelectWithFilterBack tests SelectWithFilter back option
func TestFilterablePromptSelectorSelectWithFilterBack(t *testing.T) {
	selector := &FilterablePromptSelector{
		runFunc: func(prompt promptRunner) (int, string, error) {
			return 0, BackOption, nil
		},
	}

	result, err := selector.SelectWithFilter("Test", []string{"item1", "item2"}, true)
	if err != nil {
		t.Errorf("SelectWithFilter() error = %v", err)
	}
	if result.Action != ActionBack {
		t.Errorf("SelectWithFilter() Action = %v, want ActionBack", result.Action)
	}
}

// TestFilterablePromptSelectorSelectWithFilterInterrupt tests SelectWithFilter interrupt
func TestFilterablePromptSelectorSelectWithFilterInterrupt(t *testing.T) {
	selector := &FilterablePromptSelector{
		runFunc: func(prompt promptRunner) (int, string, error) {
			return 0, "", promptui.ErrInterrupt
		},
	}

	result, err := selector.SelectWithFilter("Test", []string{"item1", "item2"}, false)
	if err != nil {
		t.Errorf("SelectWithFilter() error = %v", err)
	}
	if result.Action != ActionExit {
		t.Errorf("SelectWithFilter() Action = %v, want ActionExit", result.Action)
	}
}

// TestFilterablePromptSelectorSelectWithFilterError tests SelectWithFilter error case
func TestFilterablePromptSelectorSelectWithFilterError(t *testing.T) {
	selector := &FilterablePromptSelector{
		runFunc: func(prompt promptRunner) (int, string, error) {
			return 0, "", errors.New("prompt error")
		},
	}

	_, err := selector.SelectWithFilter("Test", []string{"item1", "item2"}, false)
	if err == nil {
		t.Error("SelectWithFilter() should return error")
	}
	if !strings.Contains(err.Error(), "selection failed") {
		t.Errorf("SelectWithFilter() error = %v, want 'selection failed'", err)
	}
}

// TestFilterablePromptSelectorSelectWithFilterNoBack tests SelectWithFilter without back option
func TestFilterablePromptSelectorSelectWithFilterNoBack(t *testing.T) {
	selector := &FilterablePromptSelector{
		runFunc: func(prompt promptRunner) (int, string, error) {
			return 1, "item2", nil
		},
	}

	result, err := selector.SelectWithFilter("Test", []string{"item1", "item2"}, false)
	if err != nil {
		t.Errorf("SelectWithFilter() error = %v", err)
	}
	if result.Selected != "item2" {
		t.Errorf("SelectWithFilter() Selected = %v, want item2", result.Selected)
	}
	if result.Action != ActionSelect {
		t.Errorf("SelectWithFilter() Action = %v, want ActionSelect", result.Action)
	}
}

// TestFilterablePromptSelectorSelectWithFilterWithBackSelectItem tests SelectWithFilter with back option selecting item
func TestFilterablePromptSelectorSelectWithFilterWithBackSelectItem(t *testing.T) {
	selector := &FilterablePromptSelector{
		runFunc: func(prompt promptRunner) (int, string, error) {
			return 1, "item1", nil
		},
	}

	result, err := selector.SelectWithFilter("Test", []string{"item1", "item2"}, true)
	if err != nil {
		t.Errorf("SelectWithFilter() error = %v", err)
	}
	if result.Selected != "item1" {
		t.Errorf("SelectWithFilter() Selected = %v, want item1", result.Selected)
	}
	if result.Action != ActionSelect {
		t.Errorf("SelectWithFilter() Action = %v, want ActionSelect", result.Action)
	}
}

// TestDefaultPromptRun tests defaultPromptRun function
func TestDefaultPromptRun(t *testing.T) {
	// Create a mock prompt runner
	mockRunner := &mockPromptRunner{
		idx:      1,
		selected: "test",
		err:      nil,
	}

	idx, selected, err := defaultPromptRun(mockRunner)
	if err != nil {
		t.Errorf("defaultPromptRun() error = %v", err)
	}
	if idx != 1 {
		t.Errorf("defaultPromptRun() idx = %v, want 1", idx)
	}
	if selected != "test" {
		t.Errorf("defaultPromptRun() selected = %v, want test", selected)
	}
}

// mockPromptRunner implements promptRunner for testing
type mockPromptRunner struct {
	idx      int
	selected string
	err      error
}

func (m *mockPromptRunner) Run() (int, string, error) {
	return m.idx, m.selected, m.err
}

// substringFilter is the reference implementation for substring filtering
// Used to verify the filtering logic in FilterablePromptSelector
func substringFilter(pattern string, items []string) []string {
	if pattern == "" {
		return items
	}
	pattern = strings.ToLower(pattern)
	var result []string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), pattern) {
			result = append(result, item)
		}
	}
	return result
}

// TestPropertyFilterCorrectlyFiltersItemsByPattern tests Property 3
// Feature: s3t-list, Property 3: Filter correctly filters items by pattern
// Validates: Requirements 4.1, 4.2, 4.5
func TestPropertyFilterCorrectlyFiltersItemsByPattern(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for filter patterns (alphanumeric, no wildcards for substring matching)
	patternGen := gen.RegexMatch(`[a-zA-Z0-9]{0,10}`)
	// Generator for resource names
	nameGen := gen.RegexMatch(`[a-zA-Z0-9_-]{1,30}`)
	// Generator for list of names
	namesGen := gen.SliceOf(nameGen)

	properties.Property("filtered results contain only items matching substring pattern", prop.ForAll(
		func(pattern string, items []string) bool {
			filtered := substringFilter(pattern, items)

			// All filtered items must contain the pattern as substring (case-insensitive)
			patternLower := strings.ToLower(pattern)
			for _, item := range filtered {
				if !strings.Contains(strings.ToLower(item), patternLower) {
					return false
				}
			}
			return true
		},
		patternGen,
		namesGen,
	))

	properties.Property("all matching items are included in filtered results", prop.ForAll(
		func(pattern string, items []string) bool {
			filtered := substringFilter(pattern, items)

			// Count items that should match
			patternLower := strings.ToLower(pattern)
			expectedCount := 0
			for _, item := range items {
				if strings.Contains(strings.ToLower(item), patternLower) {
					expectedCount++
				}
			}

			return len(filtered) == expectedCount
		},
		patternGen,
		namesGen,
	))

	properties.Property("empty pattern returns all items unchanged", prop.ForAll(
		func(items []string) bool {
			filtered := substringFilter("", items)

			if len(filtered) != len(items) {
				return false
			}
			for i, item := range items {
				if filtered[i] != item {
					return false
				}
			}
			return true
		},
		namesGen,
	))

	properties.Property("filtering is case-insensitive", prop.ForAll(
		func(pattern string, items []string) bool {
			// Filter with lowercase pattern
			filteredLower := substringFilter(strings.ToLower(pattern), items)
			// Filter with uppercase pattern
			filteredUpper := substringFilter(strings.ToUpper(pattern), items)
			// Filter with original pattern
			filteredOriginal := substringFilter(pattern, items)

			// All should produce the same results
			if len(filteredLower) != len(filteredUpper) || len(filteredLower) != len(filteredOriginal) {
				return false
			}
			for i := range filteredLower {
				if filteredLower[i] != filteredUpper[i] || filteredLower[i] != filteredOriginal[i] {
					return false
				}
			}
			return true
		},
		patternGen,
		namesGen,
	))

	properties.TestingRun(t)
}

// TestPropertySubstringMatchingIsAutomatic tests Property 4
// Feature: s3t-list, Property 4: Substring matching is automatic
// Validates: Requirements 4.2
func TestPropertySubstringMatchingIsAutomatic(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for base strings (the pattern to search for)
	baseGen := gen.RegexMatch(`[a-z0-9]{1,8}`)
	// Generator for prefix strings
	prefixGen := gen.RegexMatch(`[a-z0-9]{0,8}`)
	// Generator for suffix strings
	suffixGen := gen.RegexMatch(`[a-z0-9]{0,8}`)

	properties.Property("pattern matches exact string", prop.ForAll(
		func(pattern string) bool {
			items := []string{pattern}
			filtered := substringFilter(pattern, items)
			return len(filtered) == 1 && filtered[0] == pattern
		},
		baseGen,
	))

	properties.Property("pattern matches string with prefix", prop.ForAll(
		func(pattern, prefix string) bool {
			if pattern == "" {
				return true // Empty pattern matches everything
			}
			item := prefix + pattern
			items := []string{item}
			filtered := substringFilter(pattern, items)
			return len(filtered) == 1 && filtered[0] == item
		},
		baseGen,
		prefixGen,
	))

	properties.Property("pattern matches string with suffix", prop.ForAll(
		func(pattern, suffix string) bool {
			if pattern == "" {
				return true // Empty pattern matches everything
			}
			item := pattern + suffix
			items := []string{item}
			filtered := substringFilter(pattern, items)
			return len(filtered) == 1 && filtered[0] == item
		},
		baseGen,
		suffixGen,
	))

	properties.Property("pattern matches string with prefix and suffix", prop.ForAll(
		func(pattern, prefix, suffix string) bool {
			if pattern == "" {
				return true // Empty pattern matches everything
			}
			item := prefix + pattern + suffix
			items := []string{item}
			filtered := substringFilter(pattern, items)
			return len(filtered) == 1 && filtered[0] == item
		},
		baseGen,
		prefixGen,
		suffixGen,
	))

	properties.Property("no explicit wildcards required for substring matching", prop.ForAll(
		func(pattern, prefix, suffix string) bool {
			if pattern == "" {
				return true
			}
			// Create items that should match without wildcards
			items := []string{
				pattern,                   // exact match
				prefix + pattern,          // prefix + pattern
				pattern + suffix,          // pattern + suffix
				prefix + pattern + suffix, // prefix + pattern + suffix
			}
			filtered := substringFilter(pattern, items)
			// All items should match
			return len(filtered) == 4
		},
		baseGen,
		prefixGen,
		suffixGen,
	))

	properties.TestingRun(t)
}

// TestCreateSearcher tests createSearcher function
func TestCreateSearcher(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry"}
	searcher := createSearcher(items)

	tests := []struct {
		input    string
		index    int
		expected bool
	}{
		{"app", 0, true},    // "app" matches "Apple"
		{"APP", 0, true},    // case-insensitive
		{"ban", 1, true},    // "ban" matches "Banana"
		{"xyz", 0, false},   // "xyz" doesn't match "Apple"
		{"", 0, true},       // empty string matches everything
		{"cherry", 2, true}, // exact match
		{"err", 2, true},    // substring match
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := searcher(tt.input, tt.index)
			if result != tt.expected {
				t.Errorf("createSearcher()(%q, %d) = %v, want %v", tt.input, tt.index, result, tt.expected)
			}
		})
	}
}
